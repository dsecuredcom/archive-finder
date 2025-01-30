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
			chunk = chunk[:0]
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

// processHostsChunk processes a *list of hosts* with concurrency.
// For each chunk, we generate archive URLs and do a single-line progress bar.
func processHostsChunk(hosts []string, config *Config, client *http.Client) error {
	// Gather all the archives that need to be checked in this chunk.
	var allArchives []string
	for _, host := range hosts {
		archives := GenerateArchivePaths(host, config.DisableDynamicEntries)
		allArchives = append(allArchives, archives...)
	}

	totalInChunk := int64(len(allArchives))
	if totalInChunk == 0 {
		return nil
	}

	// Print how many tasks are in this chunk (optional).
	fmt.Printf("\nProcessing chunk of %d hosts with %d archive checks...\n", len(hosts), totalInChunk)

	// We'll use a semaphore to cap concurrency, and a WaitGroup to wait for all tasks.
	sem := make(chan struct{}, config.Concurrency)
	var wg sync.WaitGroup

	wg.Add(int(totalInChunk)) // total archives to check
	var completed int64       // atomic counter for completed jobs

	// Start a goroutine to display a single-line progress bar.
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(300 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				c := atomic.LoadInt64(&completed)
				percent := float64(c) / float64(totalInChunk) * 100.0
				// Use "\r" to overwrite the same line.
				fmt.Printf("\rProgress: %d/%d (%.1f%%)", c, totalInChunk, percent)
			}
		}
	}()

	// Spawn workers for each archive URL in this chunk.
	for _, archiveURL := range allArchives {
		sem <- struct{}{} // occupy a "slot"
		go func(url string) {
			defer wg.Done()
			defer func() { <-sem }() // release slot

			// Actual HTTP check logic
			CheckArchive(url, client, config.Verbose)

			// Increment the completed count
			atomic.AddInt64(&completed, 1)
		}(archiveURL)
	}

	// Wait for all archives in this chunk to finish
	wg.Wait()

	// Signal the progress goroutine to exit, then finalize the display.
	close(done)
	fmt.Printf("\rProgress: %d/%d (100.0%%)\n", totalInChunk, totalInChunk)

	return nil
}
