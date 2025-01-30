package main

import (
	"fmt"
	"os"

	"github.com/dsecuredcom/archive-finder/src"
)

func main() {
	// Parse CLI flags into a config struct
	config := src.ParseFlags()

	// Create the HTTP client
	client := src.NewHTTPClient(config)

	// Process the hosts file
	if err := src.ProcessHostsFile(config, client); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing hosts file: %v\n", err)
		os.Exit(1)
	}
}
