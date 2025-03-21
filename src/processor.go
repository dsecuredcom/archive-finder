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

	basePaths, extensions, backupFolders := GetBasePathsAndExtensions(config)

	numBasePaths := len(basePaths)
	numExtensions := len(extensions)
	numBackupFolders := len(backupFolders)
	numHosts := len(lines)

	// Calculate estimated requests based on configuration
	var estimated int

	if config.OnlyDynamicEntries {
		// For dynamic entries only:
		dynamicEstimated := 0

		// Calculate module-specific estimates
		if config.ModuleDomainParts {
			// Estimate average domain parts per host (conservatively)
			avgDomainParts := 5 // Conservative estimate of domain parts per host
			// Each part generates a direct request and requests in backup folders
			domainPartRequests := avgDomainParts * numExtensions * (1 + numBackupFolders)
			dynamicEstimated += domainPartRequests
		}

		if config.ModuleFirstChars {
			// First 3 chars and first 4 chars = 2 patterns per host
			firstCharsRequests := 2 * numExtensions
			dynamicEstimated += firstCharsRequests
		}

		if config.ModuleYears {
			// 4 year-based patterns per host (backup{year}, backup-{year}, backup_{year}, backups/backup{year})
			yearBasedRequests := 4 * numExtensions
			dynamicEstimated += yearBasedRequests
		}

		if config.ModuleDate {
			// 2 date-based patterns per host (backup-{date}, backups/backup-{date})
			dateBasedRequests := 2 * numExtensions
			dynamicEstimated += dateBasedRequests
		}

		estimated = numHosts * dynamicEstimated
		PrintWithTime("Dynamic list: approximately %d requests (only dynamic entries)\n", estimated)
	} else if config.DisableDynamicEntries {
		// Only static entries
		staticEstimated := numBasePaths * numExtensions * numHosts
		estimated = staticEstimated
		PrintWithTime("Static list: %d requests (no dynamic entries)\n", estimated)
	} else {
		// Both static and dynamic entries
		staticEstimated := numBasePaths * numExtensions * numHosts

		// Dynamic entries estimation
		dynamicEstimated := 0

		if config.ModuleDomainParts {
			avgDomainParts := 5 // Conservative estimate
			domainPartRequests := avgDomainParts * numExtensions * (1 + numBackupFolders)
			dynamicEstimated += domainPartRequests
		}

		if config.ModuleFirstChars {
			firstCharsRequests := 2 * numExtensions
			dynamicEstimated += firstCharsRequests
		}

		if config.ModuleYears {
			yearBasedRequests := 4 * numExtensions
			dynamicEstimated += yearBasedRequests
		}

		if config.ModuleDate {
			dateBasedRequests := 2 * numExtensions
			dynamicEstimated += dateBasedRequests
		}

		estimated = staticEstimated + (numHosts * dynamicEstimated)
		PrintWithTime("Complete list: approximately %d requests (static + dynamic entries)\n", estimated)
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
