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
	ModuleYears           bool
	ModuleDate            bool
	ModuleDomainParts     bool
	ModuleFirstChars      bool
	BackupFolders         []string
	FetchHtmlFolders      bool
}

func ParseFlags() *Config {
	var wordList string
	var extensionList string
	var backupFolders string

	config := &Config{}
	flag.StringVar(&config.HostsFile, "hosts", "", "Path to hosts list file")
	flag.DurationVar(&config.Timeout, "timeout", 60*time.Second, "Timeout for HTTP requests")
	flag.IntVar(&config.Concurrency, "concurrency", 2500, "Maximum number of concurrent requests")
	flag.IntVar(&config.ChunkSize, "chunksize", 500, "Chunksize for internal processing")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose output")
	flag.StringVar(&config.Intensity, "intensity", "medium", "Choose scanning intensity: small, medium, or big")
	flag.StringVar(&wordList, "words", "", "Comma-separated list of words (overwrites intensity-based words)")
	flag.StringVar(&extensionList, "extensions", "", "Comma-separated list of extensions (overwrites intensity-based extensions)")
	flag.StringVar(&backupFolders, "backup-folders", "", "Comma-separated list of backup folders (overwrites intensity-based folders)")
	flag.BoolVar(&config.UseFastHTTP, "fasthttp", false, "Use fasthttp instead of net/http")
	flag.BoolVar(&config.OnlyDynamicEntries, "only-dynamic-entries", false, "Use only dynamically generated entries")
	flag.BoolVar(&config.ModuleYears, "with-year", false, "Generate based on current year")
	flag.BoolVar(&config.ModuleFirstChars, "with-first-chars", false, "Generate based on first 3-4 chars of first subdomain part")
	flag.BoolVar(&config.ModuleDate, "with-date", false, "Generate based on current date")
	flag.BoolVar(&config.ModuleDomainParts, "with-host-parts", false, "Generate based on host parts")
	flag.BoolVar(&config.FetchHtmlFolders, "with-fetch-html", false, "Extract folders from HTML content")

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

	if backupFolders != "" {
		config.BackupFolders = strings.Split(backupFolders, ",")
	}

	config.FoundHosts = make(map[string]bool)

	return config
}
