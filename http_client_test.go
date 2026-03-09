package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPerformRequest(t *testing.T) {
	// Create a test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test-Header") != "present" {
			t.Errorf("Expected X-Test-Header to be 'present'")
		}
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got %s", r.Method)
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response_success"))
	}))
	defer ts.Close()

	opts := RequestOpts{
		Method:  "POST",
		URL:     ts.URL,
		Headers: "X-Test-Header: present\nIgnoredHeader: ",
		Body:    "testbody",
	}

	result := performRequest(opts)

	if result.Error != nil {
		t.Fatalf("Expected no error, got %v", result.Error)
	}

	if result.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", result.StatusCode)
	}

	if result.Body != "response_success" {
		t.Errorf("Expected 'response_success', got %s", result.Body)
	}
}

func TestPerformRequest_InvalidURL(t *testing.T) {
	opts := RequestOpts{
		Method:  "GET",
		URL:     "://invalid-url",
		Headers: "",
		Body:    "",
	}

	result := performRequest(opts)

	if result.Error == nil {
		t.Fatal("Expected error for invalid URL, got nil")
	}
}
