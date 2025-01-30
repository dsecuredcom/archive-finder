// processor.go
package src

import (
	"bufio"
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

	scanner := bufio.NewScanner(file)
	chunk := make([]string, 0, config.ChunkSize)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		chunk = append(chunk, line)

		if len(chunk) >= config.ChunkSize {
			if err := processHostsChunk(chunk, config, client); err != nil {
				return err
			}
			// Explicitly clear the chunk slice
			chunk = make([]string, 0, config.ChunkSize)
			// Force garbage collection after each chunk
			runtime.GC()
		}
	}

	// Process remaining hosts
	if len(chunk) > 0 {
		if err := processHostsChunk(chunk, config, client); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func processHostsChunk(hosts []string, config *Config, client *http.Client) error {
	var wg sync.WaitGroup
	sem := make(chan struct{}, config.Concurrency)

	for _, host := range hosts {
		archiveChan := GenerateArchivePaths(host, config.DisableDynamicEntries)

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
					CheckArchive(url, client, config.Verbose)
				}(archiveURL)
			}
		}(archiveChan)
	}

	wg.Wait()
	return nil
}
