package src

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	basePaths = []string{
		"backup",
		"backups",
		"_backup",
		"backup1",
		"www",
		"http",
		"db",
		"dump",
		"database",
		"backup/backup",
		"backups/backup",
		"backups/http",
	}

	extensions = []string{
		"zip",
		"tar",
		"rar",
		"tar.gz",
		"7z",
		"gz",
	}

	magicBytes = map[string][]byte{
		"zip":    {0x50, 0x4B, 0x03, 0x04},
		"rar":    {0x52, 0x61, 0x72, 0x21},
		"tar":    {0x75, 0x73, 0x74, 0x61, 0x72, 0x00},
		"tar.gz": {0x1F, 0x8B},
		"7z":     {0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C},
		"gz":     {0x1F, 0x8B},
	}
)

func GenerateArchivePaths(host string, disableDynamicEntries bool) <-chan string {
	archiveChan := make(chan string, 100) // Buffered channel for some throughput

	go func() {
		defer close(archiveChan)

		baseURL := normalizeHost(host)
		if baseURL == "" {
			return
		}

		// Generate from basePaths + extensions
		for _, basePath := range basePaths {
			for _, ext := range extensions {
				archive := fmt.Sprintf("%s/%s.%s", baseURL, basePath, ext)
				archiveChan <- archive
			}
		}

		// Generate dynamic entries from subdomain parts
		if !disableDynamicEntries {
			parts := generateDomainParts(baseURL)
			for _, part := range parts {
				for _, ext := range extensions {
					archive := fmt.Sprintf("%s/%s.%s", baseURL, part, ext)
					archiveChan <- archive
				}
			}
		}

		// Date-based entries
		now := time.Now()
		currentYear := now.Year()
		todayStr := now.Format("2006-01-02")

		for _, ext := range extensions {
			archiveChan <- fmt.Sprintf("%s/backup%d.%s", baseURL, currentYear, ext)
			archiveChan <- fmt.Sprintf("%s/backup-%s.%s", baseURL, todayStr, ext)
		}
	}()

	return archiveChan
}

// CheckArchive performs the HTTP request and checks if the response looks like a valid archive.
func CheckArchive(archiveURL string, client *http.Client, verbose bool) {
	startTime := time.Now()
	resp, err := client.Get(archiveURL)

	if err != nil {
		if verbose {
			fmt.Printf("[ERROR] Request failed for %s: %v\n", archiveURL, err)
		}
		return
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)
	if verbose {
		sizeStr := "unknown"
		statusCode := fmt.Sprintf("%d", resp.StatusCode)
		if resp.ContentLength >= 0 {
			sizeStr = fmt.Sprintf("%d", resp.ContentLength)
		}
		fmt.Printf("[VERBOSE] start=%s url=%s took=%v status=%s size=%s\n",
			startTime.Format(time.RFC3339),
			archiveURL,
			duration,
			statusCode,
			sizeStr,
		)
	}

	// Only proceed if status == 200
	if resp.StatusCode == http.StatusOK {
		if verifyFromResponse(resp, archiveURL) {
			fmt.Fprintf(
				// Print on its own line to avoid overwriting the progress display
				// (especially if you run them concurrently).
				// You can also replace with log.Println if you prefer.
				// or consider some concurrency-safe logging strategy.
				// For now we just print to stdout:
				//
				// There's a small glitch with the ChatGPT formatting, let's fix it:
				os.Stdout,
				"Found archive: %s\n",
				archiveURL,
			)
		}
	}
}

// verifyFromResponse checks the first few bytes to confirm if the file is a valid archive.
func verifyFromResponse(resp *http.Response, archiveURL string) bool {
	ext := getExtension(archiveURL)

	// Special case for tar: read enough bytes to check the "ustar" signature
	if ext == "tar" {
		header := make([]byte, 512)
		if _, err := io.ReadFull(resp.Body, header); err != nil {
			return false
		}
		return bytes.Equal(header[257:262], []byte("ustar")) ||
			bytes.Equal(header[257:263], []byte("ustar\000"))
	}

	// Otherwise, read first 6 bytes
	header := make([]byte, 6)
	if _, err := io.ReadFull(resp.Body, header); err != nil {
		return false
	}

	if magic, ok := magicBytes[ext]; ok {
		return bytes.HasPrefix(header, magic)
	}
	return false
}

// getExtension extracts the known extension from a URL by matching
// our list of possible extensions.
func getExtension(archiveURL string) string {
	for _, ext := range extensions {
		if strings.HasSuffix(archiveURL, "."+ext) {
			return ext
		}
	}
	return ""
}

// normalizeHost ensures the host has an http/https scheme
// and returns scheme://host
func normalizeHost(host string) string {
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}
	u, err := url.Parse(host)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s://%s", u.Scheme, u.Host)
}
