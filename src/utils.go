package src

import (
	"net"
	"net/url"
	"strings"
)

// isMD5 checks if a string is a valid 32-character MD5 hash.
func isMD5(s string) bool {
	if len(s) != 32 {
		return false
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') ||
			(r >= 'a' && r <= 'f') ||
			(r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

// isIP checks if a string is a valid IP address.
func isIP(s string) bool {
	return net.ParseIP(s) != nil
}

// dashSplitPart splits on '-' and returns unique, trimmed parts.
func dashSplitPart(s string) []string {
	splits := strings.Split(s, "-")
	results := make([]string, 0, len(splits)+1)
	seen := make(map[string]bool)
	for _, sp := range splits {
		spTrimmed := strings.TrimSpace(sp)
		if spTrimmed != "" && !seen[spTrimmed] {
			results = append(results, spTrimmed)
			seen[spTrimmed] = true
		}
	}
	// Add the full string if not already present
	if s != "" && !seen[s] {
		results = append(results, s)
		seen[s] = true
	}
	return results
}

// generateDomainParts extracts subdomain parts, skipping IP or MD5-like parts.
func generateDomainParts(baseURL string) []string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil
	}
	hostname := u.Hostname()
	if isIP(hostname) {
		return nil
	}
	parts := strings.Split(hostname, ".")
	if len(parts) < 2 {
		return nil
	}

	// If there's only domain + TLD, just parse that one part.
	if len(parts) == 2 {
		domain := parts[0]
		if !isIP(domain) && !isMD5(domain) {
			return dashSplitPart(domain)
		}
		return nil
	}

	// For subdomains, we only take up to 2 parts (you can adjust as needed)
	subdomainParts := parts[:len(parts)-2]
	if len(subdomainParts) > 2 {
		subdomainParts = subdomainParts[:2]
	}

	var results []string
	for _, sub := range subdomainParts {
		if isIP(sub) || isMD5(sub) {
			continue
		}
		splits := dashSplitPart(sub)
		results = append(results, splits...)
	}

	return results
}
