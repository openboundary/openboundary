// Copyright 2026 OpenBoundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package schema

import (
	"testing"
)

func TestAllKinds(t *testing.T) {
	kinds := AllKinds()
	expected := []Kind{KindHTTPServer, KindMiddleware, KindPostgres, KindUsecase}

	if len(kinds) != len(expected) {
		t.Errorf("AllKinds() returned %d kinds, expected %d", len(kinds), len(expected))
	}

	for i, k := range expected {
		if kinds[i] != k {
			t.Errorf("AllKinds()[%d] = %q, expected %q", i, kinds[i], k)
		}
	}
}

func TestIsValidKind(t *testing.T) {
	tests := []struct {
		name     string
		kind     Kind
		expected bool
	}{
		{"http.server is valid", KindHTTPServer, true},
		{"middleware is valid", KindMiddleware, true},
		{"postgres is valid", KindPostgres, true},
		{"usecase is valid", KindUsecase, true},
		{"unknown kind is invalid", Kind("unknown"), false},
		{"empty kind is invalid", Kind(""), false},
		{"http.server.extra is invalid", Kind("http.server.extra"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidKind(tt.kind)
			if result != tt.expected {
				t.Errorf("IsValidKind(%q) = %v, expected %v", tt.kind, result, tt.expected)
			}
		})
	}
}

func TestRegistry(t *testing.T) {
	t.Run("NewRegistry creates empty registry", func(t *testing.T) {
		reg := NewRegistry()
		if reg == nil {
			t.Fatal("NewRegistry() returned nil")
		}
		if reg.schemas == nil {
			t.Error("NewRegistry() schemas map is nil")
		}
	})

	t.Run("Register and Get schema", func(t *testing.T) {
		reg := NewRegistry()
		mock := &mockSchema{kind: KindHTTPServer}
		reg.Register(mock)

		got, ok := reg.Get(KindHTTPServer)
		if !ok {
			t.Error("Get() returned false for registered schema")
		}
		if got != mock {
			t.Error("Get() returned wrong schema")
		}
	})

	t.Run("Get returns false for unregistered kind", func(t *testing.T) {
		reg := NewRegistry()
		_, ok := reg.Get(KindHTTPServer)
		if ok {
			t.Error("Get() returned true for unregistered kind")
		}
	})

	t.Run("Register overwrites existing schema", func(t *testing.T) {
		reg := NewRegistry()
		mock1 := &mockSchema{kind: KindHTTPServer}
		mock2 := &mockSchema{kind: KindHTTPServer}

		reg.Register(mock1)
		reg.Register(mock2)

		got, _ := reg.Get(KindHTTPServer)
		if got != mock2 {
			t.Error("Register() did not overwrite existing schema")
		}
	})
}

func TestDefaultRegistry(t *testing.T) {
	reg := DefaultRegistry()
	if reg == nil {
		t.Fatal("DefaultRegistry() returned nil")
	}
	if reg.schemas == nil {
		t.Error("DefaultRegistry() schemas map is nil")
	}
}

// mockSchema implements Schema interface for testing
type mockSchema struct {
	kind      Kind
	validateErr error
}

func (m *mockSchema) Kind() Kind {
	return m.kind
}

func (m *mockSchema) Validate(spec map[string]interface{}) error {
	return m.validateErr
}
