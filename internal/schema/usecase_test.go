// Copyright 2026 OpenBoundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package schema

import (
	"testing"
)

func TestUsecaseSchema_Kind(t *testing.T) {
	s := &UsecaseSchema{}
	if s.Kind() != KindUsecase {
		t.Errorf("Kind() = %q, expected %q", s.Kind(), KindUsecase)
	}
}

func TestUsecaseSchema_Validate(t *testing.T) {
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
			name: "spec with binds_to",
			spec: map[string]interface{}{
				"binds_to": "http.server.api:POST:/users",
				"goal":     "Create a new user",
			},
			expectError: false,
		},
		{
			name: "spec with acceptance criteria",
			spec: map[string]interface{}{
				"binds_to": "http.server.api:GET:/users",
				"goal":     "List all users",
				"acceptance_criteria": []interface{}{
					"Returns paginated list",
					"Supports filtering",
				},
			},
			expectError: false,
		},
		{
			name: "spec with middleware",
			spec: map[string]interface{}{
				"binds_to":   "http.server.api:DELETE:/users/{id}",
				"goal":       "Delete a user",
				"middleware": []interface{}{"middleware.authn", "middleware.authz"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &UsecaseSchema{}
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

func TestUsecaseSchema_ImplementsSchema(t *testing.T) {
	var _ Schema = &UsecaseSchema{}
}
