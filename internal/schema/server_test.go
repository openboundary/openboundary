// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package schema

import (
	"testing"
)

func TestHTTPServerSchema_Kind(t *testing.T) {
	s := &HTTPServerSchema{}
	if s.Kind() != KindHTTPServer {
		t.Errorf("Kind() = %q, expected %q", s.Kind(), KindHTTPServer)
	}
}

func TestHTTPServerSchema_Validate(t *testing.T) {
	tests := []struct {
		name        string
		spec        map[string]interface{}
		expectError bool
	}{
		{
			name: "valid spec with port",
			spec: map[string]interface{}{
				"port": 3000,
			},
			expectError: false,
		},
		{
			name: "valid spec with all fields",
			spec: map[string]interface{}{
				"port":      3000,
				"host":      "localhost",
				"framework": "hono",
			},
			expectError: false,
		},
		{
			name:        "missing port",
			spec:        map[string]interface{}{},
			expectError: true,
		},
		{
			name: "port as string (still passes - type checking not implemented)",
			spec: map[string]interface{}{
				"port": "3000",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &HTTPServerSchema{}
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

func TestHTTPServerSchema_ImplementsSchema(t *testing.T) {
	var _ Schema = &HTTPServerSchema{}
}
