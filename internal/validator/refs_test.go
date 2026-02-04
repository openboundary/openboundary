package validator

import (
	"testing"

	"github.com/openboundary/openboundary/internal/parser"
)

func TestNewRefValidator(t *testing.T) {
	v := NewRefValidator()
	if v == nil {
		t.Fatal("NewRefValidator() returned nil")
	}
}

func TestRefValidator_Validate(t *testing.T) {
	tests := []struct {
		name       string
		spec       *parser.Spec
		wantErrors int
	}{
		{
			name: "no references",
			spec: &parser.Spec{
				Components: []parser.Component{
					{ID: "http.server.api", Kind: "http.server", Spec: map[string]interface{}{}},
				},
			},
			wantErrors: 0,
		},
		{
			name: "valid reference",
			spec: &parser.Spec{
				Components: []parser.Component{
					{ID: "postgres.primary", Kind: "postgres", Spec: map[string]interface{}{}},
					{
						ID:   "http.server.api",
						Kind: "http.server",
						Spec: map[string]interface{}{
							"$ref": "postgres.primary",
						},
					},
				},
			},
			wantErrors: 0,
		},
		{
			name: "unresolved reference",
			spec: &parser.Spec{
				Components: []parser.Component{
					{
						ID:   "http.server.api",
						Kind: "http.server",
						Spec: map[string]interface{}{
							"$ref": "nonexistent.component",
						},
					},
				},
			},
			wantErrors: 1,
		},
		{
			name: "nested reference",
			spec: &parser.Spec{
				Components: []parser.Component{
					{ID: "postgres.primary", Kind: "postgres", Spec: map[string]interface{}{}},
					{
						ID:   "http.server.api",
						Kind: "http.server",
						Spec: map[string]interface{}{
							"database": map[string]interface{}{
								"$ref": "postgres.primary",
							},
						},
					},
				},
			},
			wantErrors: 0,
		},
		{
			name: "reference in array",
			spec: &parser.Spec{
				Components: []parser.Component{
					{ID: "middleware.auth", Kind: "middleware", Spec: map[string]interface{}{}},
					{
						ID:   "http.server.api",
						Kind: "http.server",
						Spec: map[string]interface{}{
							"middleware": []interface{}{
								map[string]interface{}{
									"$ref": "middleware.auth",
								},
							},
						},
					},
				},
			},
			wantErrors: 0,
		},
		{
			name: "multiple unresolved references",
			spec: &parser.Spec{
				Components: []parser.Component{
					{
						ID:   "http.server.api",
						Kind: "http.server",
						Spec: map[string]interface{}{
							"$ref": "nonexistent.one",
							"database": map[string]interface{}{
								"$ref": "nonexistent.two",
							},
						},
					},
				},
			},
			wantErrors: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewRefValidator()
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

func TestFindRefs(t *testing.T) {
	tests := []struct {
		name     string
		spec     map[string]interface{}
		expected []string
	}{
		{
			name:     "empty spec",
			spec:     map[string]interface{}{},
			expected: nil,
		},
		{
			name: "top-level ref",
			spec: map[string]interface{}{
				"$ref": "component.name",
			},
			expected: []string{"component.name"},
		},
		{
			name: "nested ref",
			spec: map[string]interface{}{
				"database": map[string]interface{}{
					"$ref": "postgres.primary",
				},
			},
			expected: []string{"postgres.primary"},
		},
		{
			name: "ref in array",
			spec: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"$ref": "ref.one"},
					map[string]interface{}{"$ref": "ref.two"},
				},
			},
			expected: []string{"ref.one", "ref.two"},
		},
		{
			name: "non-string ref ignored",
			spec: map[string]interface{}{
				"$ref": 123,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refs := findRefs(tt.spec)

			if len(refs) != len(tt.expected) {
				t.Errorf("findRefs() returned %d refs, expected %d", len(refs), len(tt.expected))
				return
			}

			for i, ref := range tt.expected {
				if refs[i] != ref {
					t.Errorf("findRefs()[%d] = %q, expected %q", i, refs[i], ref)
				}
			}
		})
	}
}
