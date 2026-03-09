package main

import (
	"context"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
)

// Endpoint represents a single actionable route from the OpenAPI spec
type Endpoint struct {
	Method string
	Path   string
}

// DisplayName provides a formatted string for the UI ListBox
func (e Endpoint) DisplayName() string {
	return fmt.Sprintf("%-7s %s", e.Method, e.Path)
}

// parseOpenAPI reads a YAML or JSON OpenAPI 3 spec file and extracts the endpoints.
func parseOpenAPI(filePath string) ([]Endpoint, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	// Validate the document
	err = doc.Validate(context.Background())
	if err != nil {
		// Log warning but don't strictly fail, many real-world specs have validation errors
		fmt.Printf("Warning: OpenAPI validation errors: %v\n", err)
	}

	var endpoints []Endpoint

	// Iterate over paths and methods
	if doc.Paths != nil {
		for path, pathItem := range doc.Paths.Map() {
			if pathItem.Get != nil {
				endpoints = append(endpoints, Endpoint{Method: "GET", Path: path})
			}
			if pathItem.Post != nil {
				endpoints = append(endpoints, Endpoint{Method: "POST", Path: path})
			}
			if pathItem.Put != nil {
				endpoints = append(endpoints, Endpoint{Method: "PUT", Path: path})
			}
			if pathItem.Delete != nil {
				endpoints = append(endpoints, Endpoint{Method: "DELETE", Path: path})
			}
			if pathItem.Patch != nil {
				endpoints = append(endpoints, Endpoint{Method: "PATCH", Path: path})
			}
			if pathItem.Options != nil {
				endpoints = append(endpoints, Endpoint{Method: "OPTIONS", Path: path})
			}
			if pathItem.Head != nil {
				endpoints = append(endpoints, Endpoint{Method: "HEAD", Path: path})
			}
		}
	}

	return endpoints, nil
}
