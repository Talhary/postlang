package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Endpoint represents a single actionable route from the OpenAPI spec
type Endpoint struct {
	Method  string
	Path    string
	BaseURL string
	Headers string
	Body    string
}

// DisplayName provides a formatted string for the UI ListBox
func (e Endpoint) DisplayName() string {
	return fmt.Sprintf("%-7s %s", e.Method, e.Path)
}

// generateDummyJSON recursively generates a dummy JSON string from an OpenAPI schema
func generateDummyJSON(schemaRef *openapi3.SchemaRef) string {
	if schemaRef == nil || schemaRef.Value == nil {
		return ""
	}
	schema := schemaRef.Value

	if schema.Type.Is("array") && schema.Items != nil {
		itemStr := generateDummyJSON(schema.Items)
		if itemStr != "" {
			return "[\n  " + itemStr + "\n]"
		}
		return "[]"
	}

	if schema.Type.Is("object") || len(schema.Properties) > 0 {
		dummyMap := make(map[string]interface{})
		for propName, propSchema := range schema.Properties {
			if propSchema.Value != nil {
				if propSchema.Value.Example != nil {
					dummyMap[propName] = propSchema.Value.Example
				} else {
					switch {
					case propSchema.Value.Type.Is("string"):
						dummyMap[propName] = "string"
					case propSchema.Value.Type.Is("integer"), propSchema.Value.Type.Is("number"):
						dummyMap[propName] = 0
					case propSchema.Value.Type.Is("boolean"):
						dummyMap[propName] = false
					case propSchema.Value.Type.Is("array"):
						dummyMap[propName] = []interface{}{}
					default:
						dummyMap[propName] = nil
					}
				}
			}
		}

		if len(dummyMap) > 0 {
			b, err := json.MarshalIndent(dummyMap, "", "  ")
			if err == nil {
				return string(b)
			}
		}
	}

	return ""
}

// parseOpenAPI reads a YAML or JSON OpenAPI 3 spec from a reader and extracts the endpoints.
func parseOpenAPI(reader io.Reader) ([]Endpoint, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read OpenAPI spec: %w", err)
	}

	doc, err := loader.LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	// Validate the document
	err = doc.Validate(context.Background())
	if err != nil {
		fmt.Printf("Warning: OpenAPI validation errors: %v\n", err)
	}

	var endpoints []Endpoint

	baseURL := ""
	if len(doc.Servers) > 0 {
		baseURL = doc.Servers[0].URL
	}

	// Iterate over paths and methods
	if doc.Paths != nil {
		for path, pathItem := range doc.Paths.Map() {
			for method, op := range pathItem.Operations() {
				if op == nil {
					continue
				}

				ep := Endpoint{
					Method:  strings.ToUpper(method),
					Path:    path,
					BaseURL: baseURL,
				}

				var headers []string

				if op.Security != nil || doc.Security != nil {
					headers = append(headers, "Authorization: Bearer <your_token_here>")
				}

				// Extract body
				if op.RequestBody != nil && op.RequestBody.Value != nil {
					content := op.RequestBody.Value.Content
					if mt := content.Get("application/json"); mt != nil {
						headers = append(headers, "Content-Type: application/json")
						ep.Body = generateDummyJSON(mt.Schema)
					}
				}

				ep.Headers = strings.Join(headers, "\n")
				endpoints = append(endpoints, ep)
			}
		}
	}

	return endpoints, nil
}
