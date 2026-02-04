package schema

import (
	"testing"
)

func TestPostgresSchema_Kind(t *testing.T) {
	s := &PostgresSchema{}
	if s.Kind() != KindPostgres {
		t.Errorf("Kind() = %q, expected %q", s.Kind(), KindPostgres)
	}
}

func TestPostgresSchema_Validate(t *testing.T) {
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
				"provider": "drizzle",
				"schema":   "./schema.ts",
			},
			expectError: false,
		},
		{
			name: "spec with connection",
			spec: map[string]interface{}{
				"connection": "postgresql://localhost:5432/db",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PostgresSchema{}
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

func TestPostgresSchema_ImplementsSchema(t *testing.T) {
	var _ Schema = &PostgresSchema{}
}
