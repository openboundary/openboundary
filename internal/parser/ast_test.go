package parser

import (
	"testing"
)

func TestWithPosition(t *testing.T) {
	tests := []struct {
		name   string
		file   string
		line   int
		column int
	}{
		{"basic position", "test.yaml", 10, 5},
		{"zero values", "", 0, 0},
		{"large values", "/path/to/file.yaml", 9999, 9999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos := WithPosition(tt.file, tt.line, tt.column)

			if pos.File != tt.file {
				t.Errorf("WithPosition().File = %q, expected %q", pos.File, tt.file)
			}
			if pos.Line != tt.line {
				t.Errorf("WithPosition().Line = %d, expected %d", pos.Line, tt.line)
			}
			if pos.Column != tt.column {
				t.Errorf("WithPosition().Column = %d, expected %d", pos.Column, tt.column)
			}
		})
	}
}

func TestSpec_Pos(t *testing.T) {
	pos := Position{File: "spec.yaml", Line: 1, Column: 1}
	spec := &Spec{position: pos}

	result := spec.Pos()

	if result != pos {
		t.Errorf("Spec.Pos() = %+v, expected %+v", result, pos)
	}
}

func TestComponent_Pos(t *testing.T) {
	pos := Position{File: "spec.yaml", Line: 10, Column: 3}
	comp := &Component{position: pos}

	result := comp.Pos()

	if result != pos {
		t.Errorf("Component.Pos() = %+v, expected %+v", result, pos)
	}
}

func TestSpec_Fields(t *testing.T) {
	spec := &Spec{
		Version:     "0.0.1",
		Name:        "test-api",
		Description: "Test description",
		Components: []Component{
			{ID: "http.server.api", Kind: "http.server"},
		},
	}

	if spec.Version != "0.0.1" {
		t.Errorf("Spec.Version = %q, expected %q", spec.Version, "0.0.1")
	}
	if spec.Name != "test-api" {
		t.Errorf("Spec.Name = %q, expected %q", spec.Name, "test-api")
	}
	if spec.Description != "Test description" {
		t.Errorf("Spec.Description = %q, expected %q", spec.Description, "Test description")
	}
	if len(spec.Components) != 1 {
		t.Errorf("len(Spec.Components) = %d, expected %d", len(spec.Components), 1)
	}
}

func TestComponent_Fields(t *testing.T) {
	comp := &Component{
		ID:   "http.server.api",
		Kind: "http.server",
		Spec: map[string]interface{}{
			"port":      3000,
			"framework": "hono",
		},
	}

	if comp.ID != "http.server.api" {
		t.Errorf("Component.ID = %q, expected %q", comp.ID, "http.server.api")
	}
	if comp.Kind != "http.server" {
		t.Errorf("Component.Kind = %q, expected %q", comp.Kind, "http.server")
	}
	if comp.Spec["port"] != 3000 {
		t.Errorf("Component.Spec[port] = %v, expected %v", comp.Spec["port"], 3000)
	}
}
