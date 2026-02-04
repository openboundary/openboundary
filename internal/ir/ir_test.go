package ir

import (
	"testing"

	"github.com/stack-bound/stack-bound/internal/parser"
)

func TestNew(t *testing.T) {
	spec := &parser.Spec{}
	ir := New(spec)

	if ir == nil {
		t.Fatal("New() returned nil")
	}
	if ir.Spec != spec {
		t.Error("New() did not set Spec")
	}
	if ir.Components == nil {
		t.Error("New() Components is nil")
	}
	if ir.Edges == nil {
		t.Error("New() Edges is nil")
	}
	if ir.Symbols == nil {
		t.Error("New() Symbols is nil")
	}
}

func TestParseKind(t *testing.T) {
	tests := []struct {
		input    string
		expected Kind
		wantErr  bool
	}{
		{"http.server", KindHTTPServer, false},
		{"middleware", KindMiddleware, false},
		{"postgres", KindPostgres, false},
		{"usecase", KindUsecase, false},
		{"unknown", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseKind(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("ParseKind() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ParseKind() error = %v", err)
				return
			}
			if got != tt.expected {
				t.Errorf("ParseKind() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestAllKinds(t *testing.T) {
	kinds := AllKinds()
	if len(kinds) != 4 {
		t.Errorf("AllKinds() returned %d kinds, expected 4", len(kinds))
	}

	expected := map[Kind]bool{
		KindHTTPServer: true,
		KindMiddleware: true,
		KindPostgres:   true,
		KindUsecase:    true,
	}

	for _, k := range kinds {
		if !expected[k] {
			t.Errorf("AllKinds() contains unexpected kind %q", k)
		}
	}
}

func TestIsValidKind(t *testing.T) {
	tests := []struct {
		kind     Kind
		expected bool
	}{
		{KindHTTPServer, true},
		{KindMiddleware, true},
		{KindPostgres, true},
		{KindUsecase, true},
		{Kind("unknown"), false},
		{Kind(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			if got := IsValidKind(tt.kind); got != tt.expected {
				t.Errorf("IsValidKind(%q) = %v, expected %v", tt.kind, got, tt.expected)
			}
		})
	}
}

func TestEdgeTypeConstants(t *testing.T) {
	tests := []struct {
		edgeType EdgeType
		expected string
	}{
		{EdgeTypeRef, "ref"},
		{EdgeTypeDependency, "dependency"},
		{EdgeTypeMiddleware, "middleware"},
		{EdgeTypeBinding, "binding"},
	}

	for _, tt := range tests {
		if string(tt.edgeType) != tt.expected {
			t.Errorf("EdgeType %v = %q, expected %q", tt.edgeType, string(tt.edgeType), tt.expected)
		}
	}
}
