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

func ProcessHostsFile(config *Config, stdClient *http.Client, fastClient *FastHTTPClient) error {

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

	basePaths, extensions := GetBasePathsAndExtensions(config)

	numBasePaths := len(basePaths)
	numExtensions := len(extensions)
	// basePaths * extensions + 2 date-based * extensions
	knownPerHost := numBasePaths*numExtensions + 2*numExtensions
	estimated := knownPerHost * len(lines)

	if config.OnlyDynamicEntries {
		estimatedEntries := len(lines) * numExtensions * 6 // rough estimation
		PrintWithTime("Dynamic list: at least ~%d requests (only dynamic entries)\n", estimatedEntries)
	} else {
		PrintWithTime("Static list: %d requests (no dynamic subdomain entries)\n", estimated)
	}

	var chunk []string
	chunkSize := config.ChunkSize
	chunk = make([]string, 0, chunkSize)

	for _, host := range lines {
		chunk = append(chunk, host)
		if len(chunk) >= chunkSize {
			if err := processHostsChunk(chunk, config, stdClient, fastClient); err != nil {
				return err
			}
			chunk = make([]string, 0, chunkSize)
			runtime.GC()
		}
	}

	if len(chunk) > 0 {
		if err := processHostsChunk(chunk, config, stdClient, fastClient); err != nil {
			return err
		}
	}

	return nil
}

func processHostsChunk(hosts []string, config *Config, stdClient *http.Client, fastClient *FastHTTPClient) error {
	var wg sync.WaitGroup
	sem := make(chan struct{}, config.Concurrency)

	for _, host := range hosts {
		archiveChan := GenerateArchivePaths(host, config)

		wg.Add(1)
		go func(ch <-chan string) {
			defer wg.Done()
			for archiveURL := range ch {
				sem <- struct{}{}
				wg.Add(1)
				go func(url string) {
					defer wg.Done()
					defer func() { <-sem }()
					CheckArchive(url, stdClient, fastClient, config, config.Verbose)
				}(archiveURL)
			}
		}(archiveChan)
	}
	wg.Wait()
	return nil
}
