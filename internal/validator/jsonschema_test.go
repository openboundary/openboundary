// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package validator

import (
	"testing"

	"github.com/openboundary/openboundary/internal/parser"
)

func TestNewJSONSchemaValidator(t *testing.T) {
	v, err := NewJSONSchemaValidator()
	if err != nil {
		t.Fatalf("NewJSONSchemaValidator() error = %v", err)
	}
	if v == nil {
		t.Fatal("NewJSONSchemaValidator() returned nil")
	}
	if v.schema == nil {
		t.Error("NewJSONSchemaValidator() schema is nil")
	}
}

func TestJSONSchemaValidator_Validate(t *testing.T) {
	v, err := NewJSONSchemaValidator()
	if err != nil {
		t.Fatalf("NewJSONSchemaValidator() error = %v", err)
	}

	tests := []struct {
		name       string
		spec       *parser.Spec
		wantErrors bool
	}{
		{
			name: "valid minimal spec",
			spec: &parser.Spec{
				Version: "0.0.1",
				Name:    "test-api",
				Components: []parser.Component{
					{
						ID:   "http.server.api",
						Kind: "http.server",
						Spec: map[string]interface{}{
							"framework": "hono",
							"port":      3000,
						},
					},
				},
			},
			wantErrors: false,
		},
		{
			name: "valid spec with all component types",
			spec: &parser.Spec{
				Version:     "0.0.1",
				Name:        "full-api",
				Description: "A full API",
				Components: []parser.Component{
					{
						ID:   "http.server.api",
						Kind: "http.server",
						Spec: map[string]interface{}{
							"framework": "hono",
							"port":      3000,
						},
					},
					{
						ID:   "postgres.primary",
						Kind: "postgres",
						Spec: map[string]interface{}{
							"provider": "drizzle",
							"schema":   "./schema.ts",
						},
					},
					{
						ID:   "middleware.auth",
						Kind: "middleware",
						Spec: map[string]interface{}{
							"provider": "better-auth",
							"config":   "./auth.config.ts",
						},
					},
					{
						ID:   "usecase.create-user",
						Kind: "usecase",
						Spec: map[string]interface{}{
							"binds_to": "http.server.api:POST:/users",
							"goal":     "Create a user",
						},
					},
				},
			},
			wantErrors: false,
		},
		{
			name: "invalid version",
			spec: &parser.Spec{
				Version:    "",
				Name:       "test-api",
				Components: []parser.Component{},
			},
			wantErrors: true,
		},
		{
			name: "invalid component ID format",
			spec: &parser.Spec{
				Version: "0.0.1",
				Name:    "test-api",
				Components: []parser.Component{
					{
						ID:   "invalid",
						Kind: "http.server",
						Spec: map[string]interface{}{},
					},
				},
			},
			wantErrors: true,
		},
		{
			name: "invalid component kind",
			spec: &parser.Spec{
				Version: "0.0.1",
				Name:    "test-api",
				Components: []parser.Component{
					{
						ID:   "test.component",
						Kind: "invalid.kind",
						Spec: map[string]interface{}{},
					},
				},
			},
			wantErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := v.Validate(tt.spec)
			hasErrors := len(errs) > 0

			if hasErrors != tt.wantErrors {
				t.Errorf("Validate() hasErrors = %v, wantErrors = %v", hasErrors, tt.wantErrors)
				for _, e := range errs {
					t.Logf("  error: %v", e)
				}
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ValidationError
		expected string
	}{
		{
			name: "error with path",
			err: ValidationError{
				Message: "invalid value",
				Path:    "/components/0/id",
			},
			expected: "invalid value (at /components/0/id)",
		},
		{
			name: "error without path",
			err: ValidationError{
				Message: "invalid value",
			},
			expected: "invalid value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestConvertComponents(t *testing.T) {
	components := []parser.Component{
		{
			ID:   "http.server.api",
			Kind: "http.server",
			Spec: map[string]interface{}{"port": 3000},
		},
		{
			ID:   "postgres.primary",
			Kind: "postgres",
			Spec: map[string]interface{}{"provider": "drizzle"},
		},
	}

	result := convertComponents(components)

	if len(result) != 2 {
		t.Fatalf("convertComponents() returned %d items, expected 2", len(result))
	}

	if result[0]["id"] != "http.server.api" {
		t.Errorf("result[0][id] = %v, expected %q", result[0]["id"], "http.server.api")
	}
	if result[0]["kind"] != "http.server" {
		t.Errorf("result[0][kind] = %v, expected %q", result[0]["kind"], "http.server")
	}
	if result[1]["id"] != "postgres.primary" {
		t.Errorf("result[1][id] = %v, expected %q", result[1]["id"], "postgres.primary")
	}
}

func TestConvertComponents_Empty(t *testing.T) {
	result := convertComponents([]parser.Component{})
	if len(result) != 0 {
		t.Errorf("convertComponents([]) returned %d items, expected 0", len(result))
	}
}

func TestJSONSchemaValidator_Validate_MultipleErrors(t *testing.T) {
	v, _ := NewJSONSchemaValidator()

	// Spec with multiple validation errors
	spec := &parser.Spec{
		Version: "invalid", // Invalid version format (must be semver)
		Name:    "test-api",
		Components: []parser.Component{
			{
				ID:   "invalid", // Invalid ID format
				Kind: "http.server",
				Spec: map[string]interface{}{},
			},
		},
	}

	errs := v.Validate(spec)
	if len(errs) == 0 {
		t.Error("Validate() expected errors for invalid spec")
	}
}

func TestJSONSchemaValidator_Validate_ExtractsPath(t *testing.T) {
	v, _ := NewJSONSchemaValidator()

	// Spec with error that should have path extracted
	spec := &parser.Spec{
		Version: "0.0.1",
		Name:    "test-api",
		Components: []parser.Component{
			{
				ID:   "http.server.api",
				Kind: "http.server",
				Spec: map[string]interface{}{
					// Missing required fields
				},
			},
		},
	}

	errs := v.Validate(spec)
	if len(errs) == 0 {
		t.Error("Validate() expected errors")
		return
	}

	// Check that at least one error was extracted
	hasMessage := false
	for _, e := range errs {
		if e.Message != "" {
			hasMessage = true
		}
	}
	if !hasMessage {
		t.Error("Validate() errors should have messages")
	}
}
