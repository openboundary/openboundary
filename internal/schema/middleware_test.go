package schema

import (
	"testing"
)

func TestMiddlewareSchema_Kind(t *testing.T) {
	s := &MiddlewareSchema{}
	if s.Kind() != KindMiddleware {
		t.Errorf("Kind() = %q, expected %q", s.Kind(), KindMiddleware)
	}
}

func TestMiddlewareSchema_Validate(t *testing.T) {
	tests := []struct {
		name        string
		spec        map[string]interface{}
		expectError bool
	}{
		{
			name:        "empty spec (currently passes)",
			spec:        map[string]interface{}{},
			expectError: false,
		},
		{
			name: "spec with provider",
			spec: map[string]interface{}{
				"provider": "better-auth",
				"config":   "./auth.config.ts",
			},
			expectError: false,
		},
		{
			name: "casbin middleware",
			spec: map[string]interface{}{
				"provider": "casbin",
				"model":    "./model.conf",
				"policy":   "./policy.csv",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MiddlewareSchema{}
			err := s.Validate(tt.spec)

			if tt.expectError && err == nil {
				t.Error("Validate() expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestMiddlewareSchema_ImplementsSchema(t *testing.T) {
	var _ Schema = &MiddlewareSchema{}
}
