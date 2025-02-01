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

- `-hosts string`  
  Path to the hosts list file. **(Required)**

- `-timeout duration`  
  Timeout for HTTP requests (default 60s).

- `-concurrency int`  
  Maximum number of concurrent requests (default 2500).

- `-chunksize int`  
  Maximum number of hosts per batch (default 500).

- `-disable-dynamic-entries`  
  Disable generation of archive entries based on host's domain parts (default false).

- `-intensity`
  small, medium, big (see code) (default medium).

- `-words`
  Comma-separated list of words (overwrites intensity-based words)

- `-extensions`
  Comma-separated list of extensions (overwrites intensity-based extensions)

- `-fasthttp`
  Uses fasthttp to send requests instead of stlib

- `-verbose`  
  Enable verbose output (default false).

### Example

```bash
./archive-finder -hosts myhosts.txt -verbose -concurrency 1000
```

- Reads the file `myhosts.txt` (one host per line).
- Displays additional logging about each request.
- Uses a concurrency limit of 1000.

## Contributing

1. Fork this repository
2. Create a new branch: `git checkout -b feature/my-feature`
3. Make your changes and commit them
4. Push to your fork and open a pull request

## License

This project is licensed under the [MIT License](LICENSE). Feel free to use or modify for your own purposes.