// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package validator

import (
	"testing"

	"github.com/openboundary/openboundary/internal/parser"
)

func TestNewSemanticValidator(t *testing.T) {
	v := NewSemanticValidator()
	if v == nil {
		t.Fatal("NewSemanticValidator() returned nil")
	}
}

func TestSemanticValidator_Validate(t *testing.T) {
	tests := []struct {
		name       string
		spec       *parser.Spec
		wantErrors int
	}{
		{
			name: "valid spec",
			spec: &parser.Spec{
				Version: "0.0.1",
				Name:    "test-api",
				Components: []parser.Component{
					{ID: "http.server.api", Kind: "http.server"},
				},
			},
			wantErrors: 0,
		},
		{
			name: "missing version",
			spec: &parser.Spec{
				Name:       "test-api",
				Components: []parser.Component{},
			},
			wantErrors: 1,
		},
		{
			name: "missing name",
			spec: &parser.Spec{
				Version:    "0.0.1",
				Components: []parser.Component{},
			},
			wantErrors: 1,
		},
		{
			name: "missing both version and name",
			spec: &parser.Spec{
				Components: []parser.Component{},
			},
			wantErrors: 2,
		},
		{
			name: "duplicate component IDs",
			spec: &parser.Spec{
				Version: "0.0.1",
				Name:    "test-api",
				Components: []parser.Component{
					{ID: "http.server.api", Kind: "http.server"},
					{ID: "http.server.api", Kind: "http.server"},
				},
			},
			wantErrors: 1,
		},
		{
			name: "multiple duplicate IDs",
			spec: &parser.Spec{
				Version: "0.0.1",
				Name:    "test-api",
				Components: []parser.Component{
					{ID: "comp.one", Kind: "http.server"},
					{ID: "comp.one", Kind: "http.server"},
					{ID: "comp.two", Kind: "postgres"},
					{ID: "comp.two", Kind: "postgres"},
				},
			},
			wantErrors: 2,
		},
		{
			name: "empty components allowed",
			spec: &parser.Spec{
				Version:    "0.0.1",
				Name:       "test-api",
				Components: []parser.Component{},
			},
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewSemanticValidator()
			errs := v.Validate(tt.spec)

			if len(errs) != tt.wantErrors {
				t.Errorf("Validate() returned %d errors, expected %d", len(errs), tt.wantErrors)
				for _, err := range errs {
					t.Logf("  error: %v", err)
				}
			}
		})
	}
}

func TestSemanticValidator_DuplicateErrorFormat(t *testing.T) {
	spec := &parser.Spec{
		Version: "0.0.1",
		Name:    "test-api",
		Components: []parser.Component{
			{ID: "http.server.api", Kind: "http.server"},
			{ID: "http.server.api", Kind: "http.server"},
		},
	}

	v := NewSemanticValidator()
	errs := v.Validate(spec)

	if len(errs) == 0 {
		t.Fatal("expected error for duplicate component")
	}

	errStr := errs[0].Error()
	if errStr == "" {
		t.Error("error message is empty")
	}
}
