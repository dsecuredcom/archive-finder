package main

import (
	"os"
	"sync/atomic"
	"time"

	"github.com/dsecuredcom/archive-finder/src"
)

func main() {

	src.PrintWithTime("Starting archive-finder...")

	config := src.ParseFlags()

	client := src.NewHTTPClient(config)

	stopProgress := make(chan struct{})
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stopProgress:
				return
			case <-ticker.C:
				done := atomic.LoadInt64(&config.CompletedRequests)
				src.PrintProgressLine("Requests completed: %d", done)
			}
		}
	}()

	err := src.ProcessHostsFile(config, client)

	close(stopProgress)

	done := atomic.LoadInt64(&config.CompletedRequests)
	src.PrintWithTime("All done! Total requests: %d", done)

	if err != nil {
		src.PrintError("Error processing hosts file: %v", err)
		os.Exit(1)
	}
}
