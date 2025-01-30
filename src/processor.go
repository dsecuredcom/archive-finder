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
	// 1) Read entire file and collect lines (hosts).
	file, err := os.Open(config.HostsFile)
	if err != nil {
		return err
	}
	defer file.Close()

	var hosts []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		host := strings.TrimSpace(scanner.Text())
		if host != "" {
			hosts = append(hosts, host)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// 2) From these hosts, generate all possible archive URLs.
	var allArchives []string
	for _, host := range hosts {
		archives := GenerateArchivePaths(host, config.DisableDynamicEntries)
		allArchives = append(allArchives, archives...)
	}

	totalJobs := int64(len(allArchives))
	if totalJobs == 0 {
		fmt.Println("No archives to check.")
		return nil
	}

	// 3) We'll process everything with concurrency.
	//    Create a channel to feed archive URLs.
	jobs := make(chan string, config.Concurrency*2)

	var wg sync.WaitGroup
	var processed int64

	// 4) Launch worker goroutines.
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

	// 5) Launch a goroutine to display the progress bar on a single line.
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				current := atomic.LoadInt64(&processed)
				percent := float64(current) / float64(totalJobs) * 100
				// \r returns the cursor to the start of the line
				// so we overwrite the same line each time.
				fmt.Printf("\rProgress: %d/%d (%.1f%%)", current, totalJobs, percent)
			}
		}
	}()

	// 6) Feed the jobs channel.
	for _, archiveURL := range allArchives {
		jobs <- archiveURL
	}

	// 7) Close the jobs channel and wait for all workers.
	close(jobs)
	wg.Wait()

	// 8) Stop the progress ticker, print a final newline.
	close(done)
	fmt.Printf("\rProgress: %d/%d (100.0%%)\n", totalJobs, totalJobs)

	return nil
}
