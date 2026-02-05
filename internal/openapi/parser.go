// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package openapi

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Parser parses OpenAPI specification files.
type Parser struct {
	baseDir string
}

// NewParser creates a new OpenAPI parser.
// baseDir is used to resolve relative file paths.
func NewParser(baseDir string) *Parser {
	return &Parser{baseDir: baseDir}
}

// ParseFile parses an OpenAPI file and returns a Document.
func (p *Parser) ParseFile(filename string) (*Document, error) {
	// Resolve relative path
	path := filename
	if !filepath.IsAbs(filename) {
		path = filepath.Join(p.baseDir, filename)
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	spec, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	return p.convertSpec(spec)
}

// ParseBytes parses OpenAPI content from bytes.
func (p *Parser) ParseBytes(data []byte) (*Document, error) {
	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	return p.convertSpec(spec)
}

func (p *Parser) convertSpec(spec *openapi3.T) (*Document, error) {
	doc := &Document{
		Operations: make(map[string]*Operation),
	}

	if spec.Info != nil {
		doc.Title = spec.Info.Title
		doc.Version = spec.Info.Version
	}

	// Extract operations from paths
	for path, pathItem := range spec.Paths.Map() {
		if pathItem == nil {
			continue
		}

		ops := map[string]*openapi3.Operation{
			"GET":     pathItem.Get,
			"POST":    pathItem.Post,
			"PUT":     pathItem.Put,
			"PATCH":   pathItem.Patch,
			"DELETE":  pathItem.Delete,
			"HEAD":    pathItem.Head,
			"OPTIONS": pathItem.Options,
		}

		for method, op := range ops {
			if op == nil {
				continue
			}

			operation := p.convertOperation(method, path, op)
			key := operation.OperationKey()
			doc.Operations[key] = operation
		}
	}

	return doc, nil
}

func (p *Parser) convertOperation(method, path string, op *openapi3.Operation) *Operation {
	operation := &Operation{
		OperationID: op.OperationID,
		Method:      method,
		Path:        path,
		Summary:     op.Summary,
		Description: op.Description,
		Tags:        op.Tags,
		Parameters:  []Parameter{},
		Responses:   make(map[string]*Response),
	}

	// Convert parameters
	for _, paramRef := range op.Parameters {
		if paramRef == nil || paramRef.Value == nil {
			continue
		}
		param := paramRef.Value
		operation.Parameters = append(operation.Parameters, Parameter{
			Name:        param.Name,
			In:          param.In,
			Required:    param.Required,
			Description: param.Description,
			Schema:      p.convertSchemaRef(param.Schema),
		})
	}

	// Convert request body
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		rb := op.RequestBody.Value
		operation.RequestBody = &RequestBody{
			Required:    rb.Required,
			Description: rb.Description,
			Content:     make(map[string]*MediaType),
		}
		for mediaType, content := range rb.Content {
			operation.RequestBody.Content[mediaType] = &MediaType{
				Schema: p.convertSchemaRef(content.Schema),
			}
		}
	}

	// Convert responses
	if op.Responses != nil {
		for status, respRef := range op.Responses.Map() {
			if respRef == nil || respRef.Value == nil {
				continue
			}
			resp := respRef.Value
			description := ""
			if resp.Description != nil {
				description = *resp.Description
			}
			response := &Response{
				Description: description,
				Content:     make(map[string]*MediaType),
			}
			for mediaType, content := range resp.Content {
				response.Content[mediaType] = &MediaType{
					Schema: p.convertSchemaRef(content.Schema),
				}
			}
			operation.Responses[status] = response
		}
	}

	return operation
}

func (p *Parser) convertSchemaRef(ref *openapi3.SchemaRef) *Schema {
	if ref == nil {
		return nil
	}

	schema := &Schema{}

	// Check for $ref
	if ref.Ref != "" {
		schema.Ref = ref.Ref
		return schema
	}

	if ref.Value == nil {
		return schema
	}

	s := ref.Value
	types := s.Type.Slice()
	if len(types) > 0 {
		schema.Type = types[0]
	}
	schema.Format = s.Format
	schema.Description = s.Description
	schema.Nullable = s.Nullable
	schema.Required = s.Required

	// Handle enum
	if len(s.Enum) > 0 {
		schema.Enum = s.Enum
	}

	// Handle object properties
	if len(s.Properties) > 0 {
		schema.Properties = make(map[string]*Schema)
		for name, propRef := range s.Properties {
			schema.Properties[name] = p.convertSchemaRef(propRef)
		}
	}

	// Handle array items
	if s.Items != nil {
		schema.Items = p.convertSchemaRef(s.Items)
	}

	return schema
}

// ParseBinding parses a binds_to value into server ID, method, and path.
// Format: server-id:METHOD:/path
func ParseBinding(bindsTo string) (serverID, method, path string, err error) {
	if bindsTo == "" {
		return "", "", "", fmt.Errorf("empty binds_to value")
	}

	// Find first colon (after server ID)
	firstColon := strings.Index(bindsTo, ":")
	if firstColon == -1 {
		return "", "", "", fmt.Errorf("invalid binds_to format: %s (expected server:METHOD:/path)", bindsTo)
	}

	serverID = bindsTo[:firstColon]
	rest := bindsTo[firstColon+1:]

	// Find second colon (after method)
	secondColon := strings.Index(rest, ":")
	if secondColon == -1 {
		return "", "", "", fmt.Errorf("invalid binds_to format: %s (expected server:METHOD:/path)", bindsTo)
	}

	method = rest[:secondColon]
	path = rest[secondColon+1:]

	// Validate method
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "PATCH": true,
		"DELETE": true, "HEAD": true, "OPTIONS": true,
	}
	if !validMethods[method] {
		return "", "", "", fmt.Errorf("invalid HTTP method: %s", method)
	}

	// Validate path starts with /
	if !strings.HasPrefix(path, "/") {
		return "", "", "", fmt.Errorf("path must start with /: %s", path)
	}

	return serverID, method, path, nil
}

// OperationKey creates the lookup key for an operation.
func OperationKey(method, path string) string {
	return method + ":" + path
}
