package src

import (
	"crypto/tls"
	"net/http"
)

func NewHTTPClient(config *Config) *http.Client {
	return &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        config.Concurrency,
			MaxIdleConnsPerHost: config.Concurrency,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		},
	}
}
