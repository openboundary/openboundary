package validator

import (
	"errors"
	"testing"

	"github.com/stack-bound/stack-bound/internal/parser"
	"github.com/stack-bound/stack-bound/internal/schema"
)

func TestNewSchemaValidator(t *testing.T) {
	reg := schema.NewRegistry()
	v := NewSchemaValidator(reg)
	if v == nil {
		t.Fatal("NewSchemaValidator() returned nil")
	}
	if v.registry != reg {
		t.Error("NewSchemaValidator() did not set registry")
	}
}

func TestSchemaValidator_Validate(t *testing.T) {
	tests := []struct {
		name          string
		setupRegistry func(*schema.Registry)
		spec          *parser.Spec
		wantErrors    int
	}{
		{
			name:          "empty spec with empty registry",
			setupRegistry: func(r *schema.Registry) {},
			spec: &parser.Spec{
				Components: []parser.Component{},
			},
			wantErrors: 0,
		},
		{
			name:          "unknown kind",
			setupRegistry: func(r *schema.Registry) {},
			spec: &parser.Spec{
				Components: []parser.Component{
					{ID: "unknown.comp", Kind: "unknown.kind"},
				},
			},
			wantErrors: 1,
		},
		{
			name:          "valid kind without schema",
			setupRegistry: func(r *schema.Registry) {},
			spec: &parser.Spec{
				Components: []parser.Component{
					{ID: "http.server.api", Kind: "http.server"},
				},
			},
			wantErrors: 0,
		},
		{
			name: "valid component passes schema",
			setupRegistry: func(r *schema.Registry) {
				r.Register(&mockSchemaValidator{kind: schema.KindHTTPServer, validateErr: nil})
			},
			spec: &parser.Spec{
				Components: []parser.Component{
					{ID: "http.server.api", Kind: "http.server", Spec: map[string]interface{}{}},
				},
			},
			wantErrors: 0,
		},
		{
			name: "invalid component fails schema",
			setupRegistry: func(r *schema.Registry) {
				r.Register(&mockSchemaValidator{
					kind:        schema.KindHTTPServer,
					validateErr: errors.New("missing required field: port"),
				})
			},
			spec: &parser.Spec{
				Components: []parser.Component{
					{ID: "http.server.api", Kind: "http.server", Spec: map[string]interface{}{}},
				},
			},
			wantErrors: 1,
		},
		{
			name: "mixed valid and invalid components",
			setupRegistry: func(r *schema.Registry) {
				r.Register(&mockSchemaValidator{kind: schema.KindHTTPServer, validateErr: nil})
				r.Register(&mockSchemaValidator{
					kind:        schema.KindPostgres,
					validateErr: errors.New("invalid config"),
				})
			},
			spec: &parser.Spec{
				Components: []parser.Component{
					{ID: "http.server.api", Kind: "http.server", Spec: map[string]interface{}{}},
					{ID: "postgres.primary", Kind: "postgres", Spec: map[string]interface{}{}},
				},
			},
			wantErrors: 1,
		},
		{
			name: "multiple invalid components",
			setupRegistry: func(r *schema.Registry) {
				r.Register(&mockSchemaValidator{
					kind:        schema.KindHTTPServer,
					validateErr: errors.New("error1"),
				})
				r.Register(&mockSchemaValidator{
					kind:        schema.KindPostgres,
					validateErr: errors.New("error2"),
				})
			},
			spec: &parser.Spec{
				Components: []parser.Component{
					{ID: "http.server.api", Kind: "http.server", Spec: map[string]interface{}{}},
					{ID: "postgres.primary", Kind: "postgres", Spec: map[string]interface{}{}},
				},
			},
			wantErrors: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := schema.NewRegistry()
			tt.setupRegistry(reg)

			v := NewSchemaValidator(reg)
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

// mockSchemaValidator implements schema.Schema for testing
type mockSchemaValidator struct {
	kind        schema.Kind
	validateErr error
}

func (m *mockSchemaValidator) Kind() schema.Kind {
	return m.kind
}

func (m *mockSchemaValidator) Validate(spec map[string]interface{}) error {
	return m.validateErr
}
