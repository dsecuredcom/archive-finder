package src

import (
	"math/rand"
	"net"
	"net/url"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.98 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:53.0) Gecko/20100101 Firefox/53.0",
}

func GetRandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}
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

func isIP(s string) bool {
	return net.ParseIP(s) != nil
}

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

	if s != "" && !seen[s] {
		results = append(results, s)
		seen[s] = true
	}
	return results
}

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

	if len(parts) == 2 {
		domain := parts[0]
		if !isIP(domain) && !isMD5(domain) {
			return dashSplitPart(domain)
		}
		return nil
	}

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
