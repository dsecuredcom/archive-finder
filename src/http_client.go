package src

import (
	"crypto/tls"
	"net/http"
	"time"
)

func NewHTTPClient(config *Config) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:           config.Concurrency,
		MaxIdleConnsPerHost:    config.Concurrency,
		IdleConnTimeout:        30 * time.Second, // Reduced from 90s to minimize idle connection issues
		TLSClientConfig:        &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives:      false,
		ForceAttemptHTTP2:      false,            // Disable HTTP/2 to avoid potential protocol issues
		MaxResponseHeaderBytes: 4096,             // Limit header size
		ResponseHeaderTimeout:  30 * time.Second, // Add timeout for receiving response headers
		ExpectContinueTimeout:  10 * time.Second, // Add timeout for 100-continue responses
		DisableCompression:     true,             // Disable compression to reduce complexity
	}

	return &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
