package main

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/dsecuredcom/archive-finder/src"
)

func main() {
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
				fmt.Printf("\rRequests completed: %d", done)
			}
		}
	}()

	err := src.ProcessHostsFile(config, client)

	close(stopProgress)

	done := atomic.LoadInt64(&config.CompletedRequests)
	fmt.Printf("\nAll done! Total requests: %d\n", done)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing hosts file: %v\n", err)
		os.Exit(1)
	}
}
