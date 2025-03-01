# Discord Channel Content Downloader

A nifty Go tool that scans an HTML file for Discord CDN URLs, downloads the contents, and replaces the remote URLs with local file paths.

## Features

- **Image Extraction:** Scans your HTML file and finds Discord CDN URLs using a robust regex pattern.
- **Smart Downloading:** Downloads content using a custom HTTP client and handles various URL quirks (like trailing ampersands).
- **Hash-Based Filenames:** Generates unique filenames with MD5 hashes to avoid duplicate downloads and path length issues.
- **Automatic Updates:** Replaces the original URLs in the HTML file with paths to the locally saved images.
- **Self-Creating Directory:** Automatically creates a `static` directory if it doesn't exist to store your images.

## Installation :3

1. **Prerequisites:**  
    Make sure you have [Go](https://golang.org/dl/) installed.
    Have an HTML exported copy of your Discord DM/Channel/Group <br >
    (For example use the [DiscordChatExporter](https://github.com/Tyrrrz/DiscordChatExporter/) from Tyrrrz)

2. **Clone the Repository:**  
   ```bash
   git clone https://github.com/twdtech/discord-content-downloader
   cd discord-content-downloader
   go build
   ```

3. **Usage**
    ```bash
    ./dcContentLoader [your html file]
    ```
    or
    ```powershell
    .\dcContentLoader.exe .\[your html file]
    ```
