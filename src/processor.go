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

	// We'll track total # of archives as we go if needed
	var totalCount int64

	// Concurrency-limiting semaphore
	sem := make(chan struct{}, config.Concurrency)
	// WaitGroup for the entire chunk
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

	for _, host := range hosts {
		archives := GenerateArchivePaths(host, config.DisableDynamicEntries)
		for _, archiveURL := range archives {
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

	// Wait for all goroutines in this chunk
	wg.Wait()
	close(done)
	fmt.Printf("\rProcessed: %d\n", atomic.LoadInt64(&totalCount))

	return nil
}
