package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RequestOpts contains the parameters for an HTTP request
type RequestOpts struct {
	Method  string
	URL     string
	Headers string
	Body    string
}

// ResponseData contains the parsed response
type ResponseData struct {
	StatusCode int
	StatusText string
	Body       string
	Duration   time.Duration
	Error      error
}

// performRequest executes the HTTP request and returns the parsed response
func performRequest(opts RequestOpts) ResponseData {
	start := time.Now()

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var reqBody io.Reader
	if opts.Body != "" {
		reqBody = bytes.NewBuffer([]byte(opts.Body))
	}

	req, err := http.NewRequest(opts.Method, opts.URL, reqBody)
	if err != nil {
		return ResponseData{Error: fmt.Errorf("failed to create request: %w", err)}
	}

	// Parse simple key: value headers
	lines := strings.Split(opts.Headers, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			req.Header.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return ResponseData{Error: fmt.Errorf("request failed: %w", err)}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ResponseData{Error: fmt.Errorf("failed to read response body: %w", err)}
	}

	duration := time.Since(start)

	return ResponseData{
		StatusCode: resp.StatusCode,
		StatusText: resp.Status,
		Body:       string(respBody),
		Duration:   duration,
	}
}
