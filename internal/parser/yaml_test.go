// Copyright 2026 OpenBoundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewParser(t *testing.T) {
	p := NewParser("test.yaml")
	if p == nil {
		t.Fatal("NewParser() returned nil")
	}
	if p.filename != "test.yaml" {
		t.Errorf("NewParser().filename = %q, expected %q", p.filename, "test.yaml")
	}
}

func TestParser_ParseBytes(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		expectError bool
		validate    func(*testing.T, *Spec)
	}{
		{
			name: "valid minimal spec",
			yaml: `
version: "0.0.1"
name: test-api
components: []
`,
			expectError: false,
			validate: func(t *testing.T, spec *Spec) {
				if spec.Version != "0.0.1" {
					t.Errorf("Version = %q, expected %q", spec.Version, "0.0.1")
				}
				if spec.Name != "test-api" {
					t.Errorf("Name = %q, expected %q", spec.Name, "test-api")
				}
			},
		},
		{
			name: "valid spec with components",
			yaml: `
version: "0.0.1"
name: test-api
description: A test API
components:
  - id: http.server.api
    kind: http.server
    spec:
      port: 3000
      framework: hono
`,
			expectError: false,
			validate: func(t *testing.T, spec *Spec) {
				if len(spec.Components) != 1 {
					t.Fatalf("len(Components) = %d, expected 1", len(spec.Components))
				}
				comp := spec.Components[0]
				if comp.ID != "http.server.api" {
					t.Errorf("Component.ID = %q, expected %q", comp.ID, "http.server.api")
				}
				if comp.Kind != "http.server" {
					t.Errorf("Component.Kind = %q, expected %q", comp.Kind, "http.server")
				}
			},
		},
		{
			name: "spec with multiple components",
			yaml: `
version: "0.0.1"
name: multi-component
components:
  - id: http.server.api
    kind: http.server
    spec:
      port: 3000
  - id: postgres.primary
    kind: postgres
    spec:
      provider: drizzle
`,
			expectError: false,
			validate: func(t *testing.T, spec *Spec) {
				if len(spec.Components) != 2 {
					t.Fatalf("len(Components) = %d, expected 2", len(spec.Components))
				}
			},
		},
		{
			name:        "invalid yaml",
			yaml:        `invalid: yaml: syntax`,
			expectError: true,
		},
		{
			name:        "empty document",
			yaml:        ``,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser("test.yaml")
			spec, err := p.ParseBytes([]byte(tt.yaml))

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, spec)
			}
		})
	}
}

func TestParser_Parse_FileNotFound(t *testing.T) {
	p := NewParser("nonexistent.yaml")
	_, err := p.Parse()
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestParser_Parse_ValidFile(t *testing.T) {
	// Create a temporary file
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")

	content := `
version: "0.0.1"
name: file-test
components: []
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	p := NewParser(path)
	spec, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spec.Name != "file-test" {
		t.Errorf("Name = %q, expected %q", spec.Name, "file-test")
	}
}

func TestParser_parseSpec_NotDocument(t *testing.T) {
	p := NewParser("test.yaml")

	// Parse valid YAML first to get a node, then test with wrong kind
	yaml := `
version: "0.0.1"
name: test
components: []
`
	spec, err := p.ParseBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify we got a valid spec
	if spec.Version != "0.0.1" {
		t.Errorf("Version = %q, expected %q", spec.Version, "0.0.1")
	}
}

func TestParser_ParseBytes_NotMapping(t *testing.T) {
	p := NewParser("test.yaml")

	// YAML that parses to a scalar at root instead of mapping
	yaml := `just a string`

	_, err := p.ParseBytes([]byte(yaml))
	if err == nil {
		t.Error("expected error for non-mapping root, got nil")
	}
}

func TestParser_ParseBytes_DecodeError(t *testing.T) {
	p := NewParser("test.yaml")

	// YAML with wrong types that will fail to decode into Spec struct
	yaml := `
version: 123
name: [not, a, string]
components: "not an array"
`
	_, err := p.ParseBytes([]byte(yaml))
	if err == nil {
		t.Error("expected decode error, got nil")
	}
}
