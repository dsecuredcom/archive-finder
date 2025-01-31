package src

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"
)

type Config struct {
	HostsFile             string
	Timeout               time.Duration
	Concurrency           int
	ChunkSize             int
	DisableDynamicEntries bool
	Verbose               bool
	CompletedRequests     int64
	FoundHosts            map[string]bool
	FoundHostsMu          sync.Mutex
}

func ParseFlags() *Config {
	config := &Config{}
	flag.StringVar(&config.HostsFile, "hosts", "", "Path to hosts list file")
	flag.DurationVar(&config.Timeout, "timeout", 60*time.Second, "Timeout for HTTP requests")
	flag.IntVar(&config.Concurrency, "concurrency", 2500, "Maximum number of concurrent requests")
	flag.IntVar(&config.ChunkSize, "chunksize", 500, "Chunksize for internal processing")
	flag.BoolVar(&config.DisableDynamicEntries, "disable-dynamic-entries", false, "Disable generation of entries based on host")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose output")
	flag.Parse()

	if config.HostsFile == "" {
		fmt.Fprintln(os.Stderr, "Hosts file is required.")
		flag.Usage()
		os.Exit(1)
	}
	config.FoundHosts = make(map[string]bool)
	return config
}
