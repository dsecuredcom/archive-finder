// processor.go
package src

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ProcessHostsFile processes the file in CHUNKS of hosts.
// Each chunk is processed with its own single-line progress bar.
func ProcessHostsFile(config *Config, client *http.Client) error {
	file, err := os.Open(config.HostsFile)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var chunk []string
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
			// Reset chunk.
			chunk = nil
		}
	}
	// If there's any leftover chunk, process it.
	if len(chunk) > 0 {
		if err := processHostsChunk(chunk, config, client); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func processHostsChunk(hosts []string, config *Config, client *http.Client) error {
	fmt.Printf("\nProcessing chunk of %d hosts...\n", len(hosts))

	var totalCount int64
	sem := make(chan struct{}, config.Concurrency)
	var wg sync.WaitGroup

	// Set up the progress ticker
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				c := atomic.LoadInt64(&totalCount)
				fmt.Printf("\rProcessed: %d", c)
			}
		}
	}()

	// Process each host
	for _, host := range hosts {
		// Get a channel of archive paths instead of a slice
		archiveChan := GenerateArchivePaths(host, config.DisableDynamicEntries)

		// Process archives as they come in
		for archiveURL := range archiveChan {
			wg.Add(1)
			sem <- struct{}{} // Acquire a "slot"

			go func(url string) {
				defer wg.Done()
				defer func() { <-sem }() // Release slot

				CheckArchive(url, client, config.Verbose)
				atomic.AddInt64(&totalCount, 1)
			}(archiveURL)
		}
	}

	wg.Wait()
	close(done)
	fmt.Printf("\rProcessed: %d\n", atomic.LoadInt64(&totalCount))

	return nil
}
