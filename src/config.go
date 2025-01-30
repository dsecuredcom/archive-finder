package src

import (
	"flag"
	"fmt"
	"os"
	"time"
)

type Config struct {
	HostsFile             string
	Timeout               time.Duration
	Concurrency           int
	DisableDynamicEntries bool
	Verbose               bool
}

func ParseFlags() *Config {
	config := &Config{}
	flag.StringVar(&config.HostsFile, "hosts", "", "Path to hosts list file")
	flag.DurationVar(&config.Timeout, "timeout", 60*time.Second, "Timeout for HTTP requests")
	flag.IntVar(&config.Concurrency, "concurrency", 2500, "Maximum number of concurrent requests")
	flag.BoolVar(&config.DisableDynamicEntries, "disable-dynamic-entries", false, "Disable generation of entries based on host")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose output")
	flag.Parse()

	if config.HostsFile == "" {
		fmt.Fprintln(os.Stderr, "Hosts file is required.")
		flag.Usage()
		os.Exit(1)
	}

	return config
}
