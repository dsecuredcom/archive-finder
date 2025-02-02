package src

import (
	"flag"
	"fmt"
	"os"
	"strings"
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
	Intensity             string
	UserBaseWords         []string
	UserExtensions        []string
	UseFastHTTP           bool
	OnlyDynamicEntries    bool
}

func ParseFlags() *Config {
	var wordList string
	var extensionList string
	config := &Config{}
	flag.StringVar(&config.HostsFile, "hosts", "", "Path to hosts list file")
	flag.DurationVar(&config.Timeout, "timeout", 60*time.Second, "Timeout for HTTP requests")
	flag.IntVar(&config.Concurrency, "concurrency", 2500, "Maximum number of concurrent requests")
	flag.IntVar(&config.ChunkSize, "chunksize", 500, "Chunksize for internal processing")
	flag.BoolVar(&config.DisableDynamicEntries, "disable-dynamic-entries", false, "Disable generation of entries based on host")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose output")
	flag.StringVar(&config.Intensity, "intensity", "medium", "Choose scanning intensity: small, medium, or big")
	flag.StringVar(&wordList, "words", "", "Comma-separated list of words (overwrites intensity-based words)")
	flag.StringVar(&extensionList, "extensions", "", "Comma-separated list of extensions (overwrites intensity-based extensions)")
	flag.BoolVar(&config.UseFastHTTP, "fasthttp", false, "Use fasthttp instead of net/http")
	flag.BoolVar(&config.OnlyDynamicEntries, "only-dynamic-entries", false, "Use only dynamically generated entries")
	flag.Parse()

	if config.HostsFile == "" {
		fmt.Fprintln(os.Stderr, "Hosts file is required.")
		flag.Usage()
		os.Exit(1)
	}

	if wordList != "" {
		config.UserBaseWords = strings.Split(wordList, ",")
	}

	if extensionList != "" {
		config.UserExtensions = strings.Split(extensionList, ",")
	}

	config.FoundHosts = make(map[string]bool)

	return config
}
