package src

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

var (
	extensions = []string{
		"zip",
		"tar",
		"rar",
		"tar.gz",
		"7z",
		"gz",
		"bz2",
	}

	magicBytes = map[string][]byte{
		"zip":    {0x50, 0x4B, 0x03, 0x04},
		"rar":    {0x52, 0x61, 0x72, 0x21},
		"tar":    {0x75, 0x73, 0x74, 0x61, 0x72, 0x00},
		"tar.gz": {0x1F, 0x8B},
		"7z":     {0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C},
		"gz":     {0x1F, 0x8B},
		"bz2":    {0x42, 0x5A, 0x68},
	}
)

func GenerateArchivePaths(host string, config *Config) <-chan string {
	archiveChan := make(chan string, 100) // Buffered channel for some throughput

	basePaths, extensions := GetBasePathsAndExtensions(config.Intensity)

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
		if !config.DisableDynamicEntries {
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
func CheckArchive(archiveURL string, client *http.Client, config *Config, verbose bool) {

	u, err := url.Parse(archiveURL)
	if err != nil {
		atomic.AddInt64(&config.CompletedRequests, 1)
		return
	}
	host := u.Host

	// 1) QUICK CHECK: If host is already found, skip entire request
	config.FoundHostsMu.Lock()
	alreadyFound := config.FoundHosts[host]
	config.FoundHostsMu.Unlock()

	if alreadyFound {
		atomic.AddInt64(&config.CompletedRequests, 1)
		return
	}

	startTime := time.Now()
	resp, err := client.Get(archiveURL)

	if err != nil {
		if verbose {
			PrintError("Request failed for %s: %v", archiveURL, err)
		}
		atomic.AddInt64(&config.CompletedRequests, 1)
		return
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)
	if verbose {
		sizeStr := "unknown"

		if resp.ContentLength >= 0 {
			sizeStr = fmt.Sprintf("%d", resp.ContentLength)
		}

		PrintVerbose("url=%s took=%v status=%d size=%s", archiveURL, duration, resp.StatusCode, sizeStr)
	}

	// Only proceed if status == 200
	if resp.StatusCode == http.StatusOK {
		if verifyFromResponse(resp, archiveURL) {
			config.FoundHostsMu.Lock()
			if !config.FoundHosts[host] {
				config.FoundHosts[host] = true
				config.FoundHostsMu.Unlock()

				PrintFound(archiveURL) // We print only once for this host
			} else {
				// Another goroutine set it while we were verifying
				config.FoundHostsMu.Unlock()
			}
		}
	}

	_, err = io.Copy(io.Discard, resp.Body)

	if err != nil {
		if verbose {
			PrintError("Failed to fully read response body for %s: %v", archiveURL, err)
		}
	}

	atomic.AddInt64(&config.CompletedRequests, 1)
}

// verifyFromResponse checks the first few bytes to confirm if the file is a valid archive.
func verifyFromResponse(resp *http.Response, archiveURL string) bool {

	ctype := resp.Header.Get("Content-Type")
	if strings.Contains(strings.ToLower(ctype), "text/html") {
		// Very likely not an archive, unless the server is misconfigured
		return false
	}

	ext := getExtension(archiveURL)
	const maxRead = 2048
	header := make([]byte, maxRead)

	n, err := io.ReadFull(resp.Body, header)
	if err != nil {
		// If we couldn’t read enough, fallback to your existing logic or just return false
		return false
	}
	header = header[:n] // only what was read

	// Check if the chunk looks like HTML
	lowerChunk := strings.ToLower(string(header))
	if strings.Contains(lowerChunk, "<html") || strings.Contains(lowerChunk, "<!doctype") {
		return false
	}

	// Special case for tar: read enough bytes to check the "ustar" signature
	if ext == "tar" {
		// old TAR logic
		if n < 512 {
			return false
		}
		return bytes.Equal(header[257:262], []byte("ustar")) ||
			bytes.Equal(header[257:263], []byte("ustar\000"))
	}

	// For other types, fallback to your existing magic check but on `header`
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
