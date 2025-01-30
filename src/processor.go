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

func ProcessHostsFile(config *Config, client *http.Client) error {
	jobs := make(chan string, config.Concurrency*2)
	var wg sync.WaitGroup
	var processed int64

	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for archiveURL := range jobs {
				CheckArchive(archiveURL, client, config.Verbose)
				atomic.AddInt64(&processed, 1)
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		t := time.NewTicker(1 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-t.C:
				fmt.Printf("Processed: %d | In queue: %d\n", atomic.LoadInt64(&processed), len(jobs))
			}
		}
	}()

	file, err := os.Open(config.HostsFile)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		host := strings.TrimSpace(scanner.Text())
		if host == "" {
			continue
		}
		archives := GenerateArchivePaths(host, config.DisableDynamicEntries)
		for _, a := range archives {
			jobs <- a
		}
	}

	close(jobs)
	wg.Wait()
	close(done)
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
