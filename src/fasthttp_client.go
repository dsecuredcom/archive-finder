package src

import (
	"crypto/tls"
	"io"
	"time"

	"github.com/valyala/fasthttp"
)

type FastHTTPClient struct {
	client *fasthttp.Client
}

func NewFastHTTPClient(config *Config) *FastHTTPClient {
	return &FastHTTPClient{
		client: &fasthttp.Client{
			Name:                          "fasthttp-client",
			MaxConnsPerHost:               config.Concurrency,
			MaxIdleConnDuration:           30 * time.Second,
			DisablePathNormalizing:        true,
			DisableHeaderNamesNormalizing: true,
			ReadTimeout:                   config.Timeout,
			WriteTimeout:                  config.Timeout,
			MaxResponseBodySize:           -1,
			TLSConfig:                     &tls.Config{InsecureSkipVerify: true},
		},
	}
}

func (f *FastHTTPClient) DoRequest(url string, maxBytes int64) (int, string, []byte, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	req.Header.Set("Connection", "keep-alive")
	req.Header.SetProtocol("HTTP/1.1")

	err := f.client.DoRedirects(req, resp, 0)
	if err != nil {
		return 0, "", nil, err
	}

	statusCode := resp.StatusCode()
	contentType := string(resp.Header.Peek("Content-Type"))

	stream := resp.BodyStream()
	if stream == nil {
		body := resp.Body()
		if int64(len(body)) > maxBytes {
			body = body[:maxBytes]
		}
		return statusCode, contentType, body, nil
	}

	buf := make([]byte, maxBytes)
	n, readErr := stream.Read(buf)

	if readErr != nil && readErr != io.EOF {
		return statusCode, contentType, nil, readErr
	}

	return statusCode, contentType, buf[:n], nil
}
