package src

import (
	"bytes"
	"fmt"
	"github.com/valyala/fasthttp"
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
		"dll",
		"exe",
		"xls",
		"xlsx",
	}

	magicBytes = map[string][]byte{
		"zip":    {0x50, 0x4B, 0x03, 0x04},
		"rar":    {0x52, 0x61, 0x72, 0x21},
		"tar":    {0x75, 0x73, 0x74, 0x61, 0x72, 0x00},
		"tar.gz": {0x1F, 0x8B},
		"7z":     {0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C},
		"gz":     {0x1F, 0x8B},
		"bz2":    {0x42, 0x5A, 0x68},
		"dll":    {0x4D, 0x5A},
		"exe":    {0x4D, 0x5A},
		"xls":    {0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1},
		"xlsx":   {0x50, 0x4B, 0x03, 0x04},
	}
)

func doHeadStd(archiveURL string, stdClient *http.Client) (int, string, error) {
	req, err := http.NewRequest("HEAD", archiveURL, nil)
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("User-Agent", GetRandomUserAgent())
	resp, err := stdClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	return resp.StatusCode, resp.Header.Get("Content-Type"), nil
}

func doHeadFast(archiveURL string, fastClient *FastHTTPClient) (int, string, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(archiveURL)
	req.Header.SetMethod("HEAD")
	req.Header.Set("User-Agent", GetRandomUserAgent())
	req.Header.Set("Connection", "keep-alive")
	req.Header.SetProtocol("HTTP/1.1")

	err := fastClient.client.Do(req, resp)
	if err != nil {
		return 0, "", err
	}
	return resp.StatusCode(), string(resp.Header.Peek("Content-Type")), nil
}

func GenerateArchivePaths(host string, config *Config) <-chan string {
	archiveChan := make(chan string, 250) // Buffered channel for some throughput

	basePaths, extensions, commonBackupFolders := GetBasePathsAndExtensions(config)

	go func() {
		defer close(archiveChan)
		seen := make(map[string]struct{})

		addPath := func(path string) {
			if _, ok := seen[path]; !ok {
				seen[path] = struct{}{}
				archiveChan <- path
			}
		}

		baseURL := normalizeHost(host)
		if baseURL == "" {
			return
		}

		// Generate from basePaths + extensions
		if !config.OnlyDynamicEntries {
			for _, basePath := range basePaths {
				for _, ext := range extensions {
					// Using the baseURL directly which now includes the path
					archive := fmt.Sprintf("%s%s.%s", baseURL, basePath, ext)
					addPath(archive)
				}
			}
		}

		// Generate dynamic entries from subdomain parts
		if !config.DisableDynamicEntries {
			if config.ModuleDomainParts {
				parts := generateDomainParts(baseURL)
				for _, part := range parts {
					for _, ext := range extensions {
						addPath(fmt.Sprintf("%s%s.%s", baseURL, part, ext))
						for _, backupfolder := range commonBackupFolders {
							addPath(fmt.Sprintf("%s%s/%s.%s", baseURL, backupfolder, part, ext))
						}
					}
				}
			}

			if config.ModuleFirstChars {
				relevantString1 := firstSubdomainPart(baseURL, 3)
				if relevantString1 != "" {
					for _, ext := range extensions {
						addPath(fmt.Sprintf("%s%s.%s", baseURL, relevantString1, ext))
					}
				}

				relevantString2 := firstSubdomainPart(baseURL, 4)
				if relevantString2 != "" {
					for _, ext := range extensions {
						addPath(fmt.Sprintf("%s%s.%s", baseURL, relevantString2, ext))
					}
				}
			}

			// Date-based entries
			now := time.Now()
			currentYear := now.Year()
			todayStr := now.Format("2006-01-02")

			for _, ext := range extensions {
				if config.ModuleYears {
					addPath(fmt.Sprintf("%sbackup%d.%s", baseURL, currentYear, ext))
					addPath(fmt.Sprintf("%sbackup-%d.%s", baseURL, currentYear, ext))
					addPath(fmt.Sprintf("%sbackup_%d.%s", baseURL, currentYear, ext))
					addPath(fmt.Sprintf("%sbackups/backup%d.%s", baseURL, currentYear, ext))
				}

				if config.ModuleDate {
					addPath(fmt.Sprintf("%sbackup-%s.%s", baseURL, todayStr, ext))
					addPath(fmt.Sprintf("%sbackups/backup-%s.%s", baseURL, todayStr, ext))
				}
			}
		}
	}()

	return archiveChan
}

func doRequest(archiveURL string, config *Config, stdClient *http.Client, fastClient *FastHTTPClient) (int, string, []byte, error) {
	const maxRead = 2048
	var headStatus int
	var headContentType string
	var err error

	if config.UseFastHTTP {
		headStatus, headContentType, err = doHeadFast(archiveURL, fastClient)
	} else {
		headStatus, headContentType, err = doHeadStd(archiveURL, stdClient)
	}
	if err != nil {
		return 0, "", nil, err
	}

	lc := strings.ToLower(headContentType)
	if !(headStatus == 200 || headStatus == 206) || (!(strings.Contains(lc, "application")) && !(strings.Contains(lc, "octet"))) {
		return headStatus, headContentType, nil, nil
	}

	if config.UseFastHTTP {
		return fastClient.DoRequest(archiveURL, maxRead)
	} else {
		req, err := http.NewRequest("GET", archiveURL, nil)
		if err != nil {
			return 0, "", nil, err
		}
		req.Header.Set("User-Agent", GetRandomUserAgent())

		resp, err := stdClient.Do(req)
		if err != nil {
			return 0, "", nil, err
		}
		defer resp.Body.Close()

		statusCode := resp.StatusCode
		contentType := resp.Header.Get("Content-Type")

		// Die ersten maxRead Bytes lesen
		buf := make([]byte, maxRead)
		n, _ := io.ReadFull(resp.Body, buf) // Fehler ignorieren wir hier mal
		buf = buf[:n]

		return statusCode, contentType, buf, nil
	}
}

func CheckArchive(
	archiveURL string,
	stdClient *http.Client, // net/http client
	fastClient *FastHTTPClient, // fasthttp client
	config *Config,
	verbose bool,
) {
	u, err := url.Parse(archiveURL)
	if err != nil {
		atomic.AddInt64(&config.CompletedRequests, 1)
		return
	}
	host := u.Host

	config.FoundHostsMu.Lock()
	alreadyFound := config.FoundHosts[host]
	config.FoundHostsMu.Unlock()

	if alreadyFound {
		atomic.AddInt64(&config.CompletedRequests, 1)
		return
	}

	startTime := time.Now()
	statusCode, ctype, body, err := doRequest(archiveURL, config, stdClient, fastClient)
	duration := time.Since(startTime)

	if err != nil {
		if verbose {
			PrintError("Request failed for %s: %v", archiveURL, err)
		}
		atomic.AddInt64(&config.CompletedRequests, 1)
		return
	}

	if verbose {
		sizeStr := "unknown" // fasthttp kann das nicht direkt, net/http liefert ContentLength
		PrintVerbose("url=%s took=%v status=%d size=%s", archiveURL, duration, statusCode, sizeStr)
	}

	if statusCode == 200 {
		if verifyBody(body, archiveURL, ctype) {
			config.FoundHostsMu.Lock()
			if !config.FoundHosts[host] {
				config.FoundHosts[host] = true
				config.FoundHostsMu.Unlock()

				PrintFound(archiveURL) // Nur einmal pro Host
			} else {
				config.FoundHostsMu.Unlock()
			}
		}
	}

	atomic.AddInt64(&config.CompletedRequests, 1)
}

func verifyBody(body []byte, archiveURL string, ctype string) bool {
	if strings.Contains(strings.ToLower(ctype), "text/html") {
		return false
	}

	ext := getExtension(archiveURL)
	n := len(body)
	if n == 0 {
		return false
	}

	lowerChunk := strings.ToLower(string(body))
	if strings.Contains(lowerChunk, "<html") || strings.Contains(lowerChunk, "<!doctype") {
		return false
	}

	if ext == "tar" {

		if n < 512 {
			return false
		}
		return bytes.Equal(body[257:262], []byte("ustar")) ||
			bytes.Equal(body[257:263], []byte("ustar\000"))
	}

	if magic, ok := magicBytes[ext]; ok {
		return bytes.HasPrefix(body, magic)
	}
	return false
}

func getExtension(archiveURL string) string {
	for _, ext := range extensions {
		if strings.HasSuffix(archiveURL, "."+ext) {
			return ext
		}
	}
	return ""
}

func normalizeHost(host string) string {
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}
	u, err := url.Parse(host)
	if err != nil {
		return ""
	}

	// Preserve the path if it exists
	path := u.Path
	// Ensure the path ends with a slash if it's not empty
	if path != "" && !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	return fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, path)
}
