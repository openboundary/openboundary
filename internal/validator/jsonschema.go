// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package validator provides validation for specification files.
package validator

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openboundary/openboundary/internal/parser"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed openboundary.schema.json
var schemaJSON []byte

// JSONSchemaValidator validates specifications against the openboundary JSON Schema.
type JSONSchemaValidator struct {
	schema *jsonschema.Schema
}

// NewJSONSchemaValidator creates a new JSON Schema validator.
func NewJSONSchemaValidator() (*JSONSchemaValidator, error) {
	var schemaDoc any
	if err := json.Unmarshal(schemaJSON, &schemaDoc); err != nil {
		return nil, fmt.Errorf("failed to parse schema JSON: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("openboundary.schema.json", schemaDoc); err != nil {
		return nil, fmt.Errorf("failed to add schema resource: %w", err)
	}

	schema, err := compiler.Compile("openboundary.schema.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	return &JSONSchemaValidator{schema: schema}, nil
}

// Validate validates the parsed spec against the JSON Schema.
func (v *JSONSchemaValidator) Validate(spec *parser.Spec) []ValidationError {
	// Convert spec to map for JSON Schema validation
	specMap := map[string]any{
		"version":     spec.Version,
		"name":        spec.Name,
		"description": spec.Description,
		"components":  convertComponents(spec.Components),
	}

	// Round-trip through JSON to get proper interface{} types
	// that the jsonschema library expects
	jsonBytes, err := json.Marshal(specMap)
	if err != nil {
		return []ValidationError{{
			Message:  fmt.Sprintf("failed to marshal spec: %v", err),
			Position: spec.Pos(),
		}}
	}

	var specData interface{}
	if err := json.Unmarshal(jsonBytes, &specData); err != nil {
		return []ValidationError{{
			Message:  fmt.Sprintf("failed to unmarshal spec: %v", err),
			Position: spec.Pos(),
		}}
	}

	err = v.schema.Validate(specData)
	if err == nil {
		return nil
	}

	// Convert JSON Schema errors to our ValidationError format
	return convertSchemaErrors(err, spec.Pos().File)
}

// convertComponents converts parsed components to map format for validation.
func convertComponents(components []parser.Component) []map[string]interface{} {
	result := make([]map[string]interface{}, len(components))
	for i, c := range components {
		result[i] = map[string]interface{}{
			"id":   c.ID,
			"kind": c.Kind,
			"spec": c.Spec,
		}
	}
	return result
}

// ValidationError represents a validation error with location info.
// Used by both JSON schema validation and IR semantic validation.
type ValidationError struct {
	Message  string
	ID       string          // Component ID (for IR validation)
	Path     string          // JSON/YAML path (for schema validation)
	Position parser.Position // Source location
}

func (e ValidationError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("%s: %s", e.ID, e.Message)
	}
	if e.Path != "" {
		return fmt.Sprintf("%s (at %s)", e.Message, e.Path)
	}
	return e.Message
}

// convertSchemaErrors converts jsonschema errors to ValidationErrors.
func convertSchemaErrors(err error, file string) []ValidationError {
	var errors []ValidationError

	// The v6 library returns *ValidationError which implements error
	// We need to extract the detailed error info from it
	ve, ok := err.(*jsonschema.ValidationError)
	if !ok {
		return []ValidationError{{
			Message:  err.Error(),
			Position: parser.Position{File: file},
		}}
	}

	// Extract errors from the validation error tree
	errors = extractValidationErrors(ve, file)
	return errors
}

// extractValidationErrors recursively extracts validation errors from the tree.
func extractValidationErrors(ve *jsonschema.ValidationError, file string) []ValidationError {
	var errors []ValidationError

	// Get the basic errors - the validation error implements Unwrap
	// and has a structured representation
	errStr := ve.Error()

	// Parse the error string to extract location and message
	// The format is: "jsonschema validation failed at 'path': reason"
	lines := strings.Split(errStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Extract path if present (format: at '/path/to/field')
		path := ""
		if idx := strings.Index(line, "at '"); idx != -1 {
			endIdx := strings.Index(line[idx+4:], "'")
			if endIdx != -1 {
				path = line[idx+4 : idx+4+endIdx]
			}
		}

		errors = append(errors, ValidationError{
			Message:  line,
			Path:     path,
			Position: parser.Position{File: file},
		})
	}

	// If we couldn't parse any errors, just use the full error string
	if len(errors) == 0 {
		errors = append(errors, ValidationError{
			Message:  errStr,
			Position: parser.Position{File: file},
		})
	}

	return errors
}
