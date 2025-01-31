package src

import (
	"fmt"
	"os"
	"time"
)

const (
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[32m"
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
)

func PrintWithTime(format string, a ...interface{}) {
	now := time.Now().Format(time.RFC3339) // or choose another format
	fmt.Printf("[%s] %s\n", now, fmt.Sprintf(format, a...))
}

func PrintFound(archiveURL string) {
	now := time.Now().Format(time.RFC3339)
	fmt.Fprintf(
		os.Stdout,
		"\n[%s] %sFound archive: %s%s\n",
		now,
		ColorGreen,
		archiveURL,
		ColorReset,
	)
}

// printError prints error messages in red.
func PrintError(format string, a ...interface{}) {
	now := time.Now().Format(time.RFC3339)
	// You could also prefix "[ERROR]" or something similar if you like:
	fmt.Fprintf(
		os.Stderr,
		"[%s] %sERROR:%s %s\n",
		now,
		ColorRed,
		ColorReset,
		fmt.Sprintf(format, a...),
	)
}

// printVerbose prints timestamped lines for verbose info, optionally in color.
func PrintVerbose(format string, a ...interface{}) {
	now := time.Now().Format(time.RFC3339)
	fmt.Printf("[%s] %s\n", now, fmt.Sprintf(format, a...))
}

func PrintProgressLine(format string, a ...interface{}) {
	// \r   : carriage return to move cursor back to the start of the line
	// \033[K: clear from cursor to the end of the line (ANSI escape)
	now := time.Now().Format(time.RFC3339)
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("\r\033[K[%s] %s", now, msg)
}
