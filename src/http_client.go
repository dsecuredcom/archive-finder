package src

import (
	"crypto/tls"
	"net/http"
	"time"
)

func NewHTTPClient(config *Config) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        config.Concurrency,
		MaxIdleConnsPerHost: config.Concurrency,
		IdleConnTimeout:     90 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives:   false, // Enable connection reuse
	}

	return &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
	}
}
