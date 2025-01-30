package src

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

func ProcessHostsFile(config *Config, client *http.Client) error {
	file, err := os.Open(config.HostsFile)
	if err != nil {
		return fmt.Errorf("failed to open hosts file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	const chunkSize = 750
	hosts := make([]string, 0, chunkSize)

	for scanner.Scan() {
		hosts = append(hosts, scanner.Text())
		if len(hosts) >= chunkSize {
			if err := processHostsChunk(hosts, config, client); err != nil {
				return err
			}
			hosts = hosts[:0]
		}
	}

	// Handle leftover chunk if any
	if len(hosts) > 0 {
		if err := processHostsChunk(hosts, config, client); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func processHostsChunk(hosts []string, config *Config, client *http.Client) error {
	var totalInChunk int64

	// Calculate the total number of archive checks in this chunk
	for _, host := range hosts {
		archives := GenerateArchivePaths(host, config.DisableDynamicEntries)
		totalInChunk += int64(len(archives))
	}
	if totalInChunk == 0 {
		return nil
	}

	fmt.Printf("Processing chunk with %d archive checks...\n", totalInChunk)

	sem := make(chan struct{}, config.Concurrency)
	var wg sync.WaitGroup
	wg.Add(int(totalInChunk))

	var completed int64

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				c := atomic.LoadInt64(&completed)
				percent := float64(c) / float64(totalInChunk) * 100
				fmt.Printf("\rProgress: %d/%d (%.1f%%)", c, totalInChunk, percent)
			}
		}
	}()

	for _, host := range hosts {
		archives := GenerateArchivePaths(host, config.DisableDynamicEntries)
		for _, archiveURL := range archives {
			sem <- struct{}{}
			go func(url string) {
				defer wg.Done()
				defer func() { <-sem }()
				CheckArchive(url, client, config.Verbose)
				atomic.AddInt64(&completed, 1)
			}(archiveURL)
		}
	}

	wg.Wait()
	close(done)

	fmt.Printf("\rProgress: %d/%d (100.0%%)\n", totalInChunk, totalInChunk)
	return nil
}
