// main.go
package main

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/dsecuredcom/archive-finder/src"
)

func main() {
	log.SetOutput(io.Discard)
	rand.Seed(time.Now().UnixNano())

	src.PrintWithTime("Starting archive-finder...")

	config := src.ParseFlags()

	var stdClient *http.Client
	var fastClient *src.FastHTTPClient

	if config.UseFastHTTP {
		// Falls fasthttp gewünscht:
		fastClient = src.NewFastHTTPClient(config)
	} else {
		// Falls net/http gewünscht:
		stdClient = src.NewHTTPClient(config)
	}

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

	err := src.ProcessHostsFile(config, stdClient, fastClient)
	close(stopProgress)

	done := atomic.LoadInt64(&config.CompletedRequests)
	src.PrintWithTime("All done! Total requests: %d", done)

	if err != nil {
		src.PrintError("Error processing hosts file: %v", err)
		os.Exit(1)
	}
}
