// Copyright 2026 OpenBoundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package openapi provides OpenAPI specification parsing for code generation.
package openapi

// Document represents a parsed OpenAPI document.
type Document struct {
	Title      string
	Version    string
	Operations map[string]*Operation // keyed by "METHOD:/path"
}

// Operation represents an OpenAPI operation (endpoint).
type Operation struct {
	OperationID string
	Method      string
	Path        string
	Summary     string
	Description string
	Parameters  []Parameter
	RequestBody *RequestBody
	Responses   map[string]*Response // keyed by status code
	Tags        []string
}

// OperationKey returns the lookup key for an operation (e.g., "GET:/users/{id}").
func (o *Operation) OperationKey() string {
	return o.Method + ":" + o.Path
}

// Parameter represents an OpenAPI parameter (path, query, header, cookie).
type Parameter struct {
	Name        string
	In          string // "path", "query", "header", "cookie"
	Required    bool
	Description string
	Schema      *Schema
}

// RequestBody represents an OpenAPI request body.
type RequestBody struct {
	Required    bool
	Description string
	Content     map[string]*MediaType // keyed by media type (e.g., "application/json")
}

// Response represents an OpenAPI response.
type Response struct {
	Description string
	Content     map[string]*MediaType // keyed by media type
}

// MediaType represents a media type with its schema.
type MediaType struct {
	Schema *Schema
}

// Schema represents a simplified JSON Schema for type generation.
type Schema struct {
	Type        string             // "object", "array", "string", "number", "integer", "boolean"
	Format      string             // "int32", "int64", "float", "double", "date", "date-time", etc.
	Ref         string             // $ref if this is a reference
	Properties  map[string]*Schema // for object types
	Items       *Schema            // for array types
	Required    []string           // required property names
	Enum        []interface{}      // enum values
	Description string
	Nullable    bool
}

// IsRef returns true if this schema is a $ref reference.
func (s *Schema) IsRef() bool {
	return s != nil && s.Ref != ""
}

// RefName extracts the type name from a $ref (e.g., "#/components/schemas/User" -> "User").
func (s *Schema) RefName() string {
	if !s.IsRef() {
		return ""
	}
	// Extract last part after "/"
	ref := s.Ref
	for i := len(ref) - 1; i >= 0; i-- {
		if ref[i] == '/' {
			return ref[i+1:]
		}
	}
	return ref
}
