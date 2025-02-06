package src

import (
	"golang.org/x/net/publicsuffix"
	"math/rand"
	"net/url"
	"regexp"
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
var ipRegex = regexp.MustCompile(`^(?:\d{1,3}\.){3}\d{1,3}(?::\d+)?$`)
var numberRegex = regexp.MustCompile(`^\d+$`)

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

func isIPAddress(host string) bool {
	return ipRegex.MatchString(host)
}

func dashSplitPart(s string) []string {
	splits := strings.Split(s, "-")
	for i := range splits {
		splits[i] = strings.TrimSpace(splits[i])
	}
	return splits
}

func generateDomainParts(baseURL string) []string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil
	}
	hostname := u.Hostname()
	if isIPAddress(hostname) {
		return nil
	}

	// Check if we have enough parts
	parts := strings.Split(hostname, ".")
	if len(parts) < 2 {
		return nil
	}

	// We'll accumulate everything in here, then deduplicate at the end
	var allParts []string

	// Effective TLD + 1  => example: "example.co.uk" from "sub.example.co.uk"
	domainPlusTld, err := publicsuffix.EffectiveTLDPlusOne(hostname)
	if err != nil {
		return nil
	}

	// Add that and the raw hostname
	allParts = append(allParts, domainPlusTld, hostname)

	// Now get just the TLD
	tld, _ := publicsuffix.PublicSuffix(hostname)
	domainName := strings.TrimSuffix(domainPlusTld, "."+tld)
	// If it's at least 3 chars, we add it:
	if len(domainName) > 2 {
		allParts = append(allParts, domainName)
	}

	relevantSubdomainPart := strings.TrimSuffix(hostname, "."+domainPlusTld)
	subParts := strings.Split(relevantSubdomainPart, ".")

	// Per your logic, we only take up to 2 subdomain parts
	if len(subParts) > 2 {
		subParts = subParts[:2]
	}

	// Expand dash-splits from subparts, skip IPs/MD5
	for _, sp := range subParts {
		if sp == "" || isIPAddress(sp) || isMD5(sp) {
			continue
		}

		// 1) Add the entire sub-part (e.g. "test-certificate")
		allParts = append(allParts, sp)

		// 2) Then add each dash-split piece (e.g. "test", "certificate")
		dashParts := dashSplitPart(sp)
		for _, dp := range dashParts {
			dp = strings.TrimSpace(dp)
			if dp == "" {
				continue
			}
			allParts = append(allParts, dp)
		}
	}

	// Now final pass: skip numeric, and deduplicate
	encountered := make(map[string]bool, len(allParts))
	var unique []string

	for _, part := range allParts {
		// skip purely numeric
		if numberRegex.MatchString(part) {
			continue
		}
		if !encountered[part] {
			encountered[part] = true
			unique = append(unique, part)
		}
	}

	return unique
}
