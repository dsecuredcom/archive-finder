# Archive Finder

This project scans a list of hosts and attempts to locate potential archive files (e.g., `.zip`, `.tar`, `.rar`, etc.) on those hosts. It generates likely archive URLs based on known paths, domain name parts, and date-based patterns, then checks if those URLs lead to real archives.

## Requirements

- Go 1.20 (or higher)
- A file containing a list of hosts (one per line)

## Installation

1. Clone this repository:

   ```bash
   git clone https://github.com/dsecuredcom/archive-finder
   cd archive-finder
   ```

2. Initialize or update the Go modules (if needed):

   ```bash
   go mod tidy
   ```

3. Build the binary:

   ```bash
   go build -o archive-finder
   ```

## Usage

```bash
./archive-finder -hosts /path/to/hosts_file.txt [options]
```

### CLI Flags

#### Required
- `-hosts string`  
  Path to the hosts list file. **(Required)**

#### General Settings
- `-timeout duration`  
  Timeout for HTTP requests (default 60s).
- `-concurrency int`  
  Maximum number of concurrent requests (default 2500).
- `-chunksize int`  
  Chunksize for internal processing (default 500).
- `-verbose`  
  Enable verbose output (default false).

#### HTTP Client Options
- `-fasthttp`  
  Use fasthttp instead of net/http for potentially faster requests (default false).

#### Dictionary Control
- `-intensity string`  
  Choose scanning intensity: "small", "medium", or "big" (default "medium").
  Controls the built-in wordlists, extensions, and backup folder names.
- `-words string`  
  Comma-separated list of words (overwrites intensity-based words).
- `-extensions string`  
  Comma-separated list of extensions (overwrites intensity-based extensions).
- `-backup-folders string`  
  Comma-separated list of backup folders (overwrites intensity-based folders).

#### Entry Generation Modules
- `-disable-dynamic-entries`  
  Disable generation of entries based on host (default false).
- `-only-dynamic-entries`  
  Use only dynamically generated entries (default false).
- `-with-host-parts`  
  Generate based on host parts (default false).
- `-with-first-chars`  
  Generate based on first 3-4 chars of first subdomain part (default false).
- `-with-year`  
  Generate based on current year (default false).
- `-with-date`  
  Generate based on current date (default false).

### Notes

- When using dynamic entries (default behavior or with `-only-dynamic-entries`), you must activate at least one module using the `-with-*` flags.
- You cannot use both `-disable-dynamic-entries` and `-only-dynamic-entries` together.

### Examples

```bash
# Basic scan with default settings
./archive-finder -hosts myhosts.txt

# Verbose output with limited concurrency
./archive-finder -hosts myhosts.txt -verbose -concurrency 1000

# Use only domain-based dynamic entries
./archive-finder -hosts myhosts.txt -only-dynamic-entries -with-host-parts

# Comprehensive scan with all dynamic modules
./archive-finder -hosts myhosts.txt -with-host-parts -with-first-chars -with-year -with-date

# High intensity scan with fasthttp
./archive-finder -hosts myhosts.txt -intensity big -fasthttp -with-host-parts -with-year
```

## How It Works

1. Reads host entries from the provided file
2. Generates potential archive URLs based on:
    - Static wordlists (controlled by `-intensity`)
    - Dynamic patterns from domain parts (when `-with-host-parts` is enabled)
    - First characters of subdomain (when `-with-first-chars` is enabled)
    - Year-based patterns (when `-with-year` is enabled)
    - Date-based patterns (when `-with-date` is enabled)
3. Checks each URL to determine if it contains an actual archive
4. Reports findings in real-time

## Contributing

1. Fork this repository
2. Create a new branch: `git checkout -b feature/my-feature`
3. Make your changes and commit them
4. Push to your fork and open a pull request

## License

This project is licensed under the [MIT License](LICENSE). Feel free to use or modify for your own purposes.