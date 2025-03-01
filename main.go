package main

import (
	"crypto/md5"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
)

func main() {
	// Check if HTML file is provided as argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <html-file>   :3")
		os.Exit(1)
	}
	htmlFile := os.Args[1]

	// Read HTML content from file
	data, err := os.ReadFile(htmlFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}
	content := string(data)

	// Create "static" directory if it doesn't exist
	if _, err := os.Stat("static"); os.IsNotExist(err) {
		err = os.Mkdir("static", 0755)
		if err != nil {
			fmt.Printf("Error creating static directory: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Created directory: static   :3")
	}

	// Improved regex to match Discord CDN image URLs
	// Match the URL up to any HTML tag character
	re := regexp.MustCompile(`https://cdn\.discordapp\.com/[^<>"'\s]+`)
	matches := re.FindAllString(content, -1)

	// Map to avoid duplicate downloads and track replacements
	replacements := make(map[string]string)

	// Process found URLs
	for _, rawURL := range matches {
		// Skip if already processed
		if _, exists := replacements[rawURL]; exists {
			continue
		}

		// Clean the URL from any trailing HTML but keep all query parameters
		cleanedURL := extractCompleteURL(rawURL)

		// Remove trailing ampersand if present
		if strings.HasSuffix(cleanedURL, "&") {
			cleanedURL = cleanedURL[:len(cleanedURL)-1]
		}

		// Decode HTML entities like &amp;
		decodedURL := html.UnescapeString(cleanedURL)

		// Print the full URL we're attempting to download
		fmt.Printf("Downloading: %s\n", decodedURL)

		// Try the file with different URL variants if needed
		filename, err := tryDownloadWithVariants(decodedURL)
		if err != nil {
			fmt.Printf("Error downloading %s: %v\n", decodedURL, err)
			continue
		}

		localPath := "static/" + filename
		replacements[rawURL] = localPath

		// Also map the cleaned and decoded URL to the same local path
		if cleanedURL != rawURL {
			replacements[cleanedURL] = localPath
		}
		if decodedURL != cleanedURL {
			replacements[decodedURL] = localPath
		}

		fmt.Printf("Will replace with local file: %s\n", localPath)
	}

	// Sort URLs by length (longest first) to avoid partial replacements
	var urls []string
	for url := range replacements {
		urls = append(urls, url)
	}

	for i := 0; i < len(urls); i++ {
		for j := i + 1; j < len(urls); j++ {
			if len(urls[i]) < len(urls[j]) {
				urls[i], urls[j] = urls[j], urls[i]
			}
		}
	}

	// Perform the replacements in order
	for _, url := range urls {
		content = strings.ReplaceAll(content, url, replacements[url])
		fmt.Printf("Replaced URL: %s -> %s\n", url, replacements[url])
	}

	// Write the updated HTML content back to the file
	err = os.WriteFile(htmlFile, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error writing updated file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("HTML file updated successfully!")
}

// extractCompleteURL properly extracts the complete Discord CDN URL
// keeping ALL query parameters intact
func extractCompleteURL(rawURL string) string {
	// Define common HTML markers that might terminate the URL
	htmlTerminators := []string{"<", ">", "\"", "'", "</a>", "</span>"}

	cleanURL := rawURL

	// Find the earliest HTML terminator position
	earliestPos := len(cleanURL)
	for _, marker := range htmlTerminators {
		pos := strings.Index(cleanURL, marker)
		if pos > 0 && pos < earliestPos {
			earliestPos = pos
		}
	}

	// Trim at the earliest HTML terminator
	if earliestPos < len(cleanURL) {
		cleanURL = cleanURL[:earliestPos]
	}

	return cleanURL
}

// tryDownloadWithVariants attempts to download with different URL variants
// such as with/without trailing ampersand
func tryDownloadWithVariants(imageURL string) (string, error) {
	variants := []string{imageURL}

	// Add a variant without a trailing ampersand if needed
	if strings.HasSuffix(imageURL, "&") {
		variants = append(variants, imageURL[:len(imageURL)-1])
	}

	// Try all URL variants
	var lastErr error
	for _, variant := range variants {
		filename, err := downloadImage(variant)
		if err == nil {
			return filename, nil
		}
		lastErr = err
		fmt.Printf("Failed with variant %s: %v, trying next...\n", variant, err)
	}

	return "", lastErr
}

// downloadImage downloads the image from imageURL and saves it in the "static" folder
// with a hash-based filename to prevent path length issues
func downloadImage(imageURL string) (string, error) {
	// Create a client with timeout that correctly handles redirects
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	// Perform HTTP GET request to fetch the image
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return "", err
	}

	// Add a user agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Parse the URL to get the file extension
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return "", err
	}

	// Get the file extension from the URL path
	ext := path.Ext(parsedURL.Path)
	if ext == "" {
		// If no extension found, try to determine from Content-Type
		contentType := resp.Header.Get("Content-Type")
		ext = getExtensionFromContentType(contentType)
	}

	// Generate a short filename using MD5 hash of the URL
	hasher := md5.New()
	hasher.Write([]byte(imageURL))
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	// Create a short filename with the hash and extension
	filename := hash + ext

	// Create the file in the "static" directory
	filePath := "static/" + filename
	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Write image content to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return filename, nil
}

// getExtensionFromContentType tries to determine file extension from MIME type
func getExtensionFromContentType(contentType string) string {
	switch {
	case strings.Contains(contentType, "image/jpeg"):
		return ".jpg"
	case strings.Contains(contentType, "image/png"):
		return ".png"
	case strings.Contains(contentType, "image/gif"):
		return ".gif"
	case strings.Contains(contentType, "image/webp"):
		return ".webp"
	case strings.Contains(contentType, "video/mp4"):
		return ".mp4"
	case strings.Contains(contentType, "video/quicktime"):
		return ".mov"
	case strings.Contains(contentType, "audio/mpeg"):
		return ".mp3"
	default:
		return ".bin" // Generic binary extension as fallback
	}
}
