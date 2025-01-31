// processor.go
package src

import (
	"bufio"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
)

func ProcessHostsFile(config *Config, client *http.Client) error {
	file, err := os.Open(config.HostsFile)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	rand.Shuffle(len(lines), func(i, j int) {
		lines[i], lines[j] = lines[j], lines[i]
	})

	basePaths, extensions := GetBasePathsAndExtensions(config.Intensity)

	numBasePaths := len(basePaths)
	numExtensions := len(extensions)
	// basePaths * extensions + 2 date-based * extensions
	knownPerHost := numBasePaths*numExtensions + 2*numExtensions
	estimated := knownPerHost * len(lines)

	if config.DisableDynamicEntries {
		PrintWithTime("At least %d requests (no dynamic subdomain entries)\n", estimated)
	} else {
		PrintWithTime("At least %d requests + dynamically generated entries (subdomains, dash-splits, etc.)\n", estimated)
	}

	var chunk []string
	chunkSize := config.ChunkSize
	chunk = make([]string, 0, chunkSize)

	for _, host := range lines {
		chunk = append(chunk, host)
		if len(chunk) >= chunkSize {
			if err := processHostsChunk(chunk, config, client); err != nil {
				return err
			}
			chunk = make([]string, 0, chunkSize)
			// Force garbage collection after each chunk if desired
			runtime.GC()
		}
	}

	if len(chunk) > 0 {
		if err := processHostsChunk(chunk, config, client); err != nil {
			return err
		}
	}

	return nil
}

func processHostsChunk(hosts []string, config *Config, client *http.Client) error {
	var wg sync.WaitGroup
	sem := make(chan struct{}, config.Concurrency)

	for _, host := range hosts {
		archiveChan := GenerateArchivePaths(host, config)

		// Create a separate goroutine to handle each archive channel
		wg.Add(1)
		go func(ch <-chan string) {
			defer wg.Done()
			for archiveURL := range ch { // This ensures the channel is fully drained
				sem <- struct{}{}
				wg.Add(1)
				go func(url string) {
					defer wg.Done()
					defer func() { <-sem }()
					CheckArchive(url, client, config, config.Verbose)
				}(archiveURL)
			}
		}(archiveChan)
	}

	wg.Wait()
	return nil
}
