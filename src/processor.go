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

	// concurrency-limiting channel + a single WaitGroup
	sem := make(chan struct{}, config.Concurrency)
	var wg sync.WaitGroup

	var lineCount int64

	for scanner.Scan() {
		host := strings.TrimSpace(scanner.Text())
		if host == "" {
			continue
		}

		archives := GenerateArchivePaths(host, config.DisableDynamicEntries)
		for _, archiveURL := range archives {
			wg.Add(1)
			sem <- struct{}{}

			go func(url string) {
				defer wg.Done()
				defer func() { <-sem }()
				CheckArchive(url, client, config.Verbose)
			}(archiveURL)
		}
		atomic.AddInt64(&lineCount, 1)
		if lineCount%100 == 0 {
			fmt.Printf("Processed %d lines...\n", lineCount)
		}
	}

	wg.Wait() // wait for all requests to finish

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
