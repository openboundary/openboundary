// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package codegen

import (
	"testing"

	"github.com/openboundary/openboundary/internal/ir"
	"github.com/openboundary/openboundary/internal/parser"
)

func TestNewPluginRegistry(t *testing.T) {
	r := NewPluginRegistry()
	if r == nil {
		t.Fatal("NewPluginRegistry() returned nil")
	}
}

func TestPluginRegistry_RegisterAndResolve(t *testing.T) {
	r := NewPluginRegistry()

	always := GeneratorPlugin{
		Name: "always",
		NewGenerator: func() Generator {
			return &mockGenerator{name: "always"}
		},
	}
	serverOnly := GeneratorPlugin{
		Name: "server-only",
		NewGenerator: func() Generator {
			return &mockGenerator{name: "server-only"}
		},
		Supports: []ir.Kind{ir.KindHTTPServer},
	}

	if err := r.Register(always); err != nil {
		t.Fatalf("register always error = %v", err)
	}
	if err := r.Register(serverOnly); err != nil {
		t.Fatalf("register server-only error = %v", err)
	}

	i := &ir.IR{
		Spec: &parser.Spec{Name: "test", Version: "0.0.1"},
		Components: map[string]*ir.Component{
			"http.server.api": {
				ID:   "http.server.api",
				Kind: ir.KindHTTPServer,
				HTTPServer: &ir.HTTPServerSpec{
					Framework: "hono",
					Port:      3000,
				},
			},
		},
	}

	gens, err := r.GeneratorsForIR(i)
	if err != nil {
		t.Fatalf("GeneratorsForIR() error = %v", err)
	}
	if len(gens) != 2 {
		t.Fatalf("GeneratorsForIR() len = %d, expected 2", len(gens))
	}
}

func TestPluginRegistry_FilterByKind(t *testing.T) {
	r := NewPluginRegistry()
	postgresOnly := GeneratorPlugin{
		Name: "postgres-only",
		NewGenerator: func() Generator {
			return &mockGenerator{name: "postgres-only"}
		},
		Supports: []ir.Kind{ir.KindPostgres},
	}

	if err := r.Register(postgresOnly); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	i := &ir.IR{
		Spec: &parser.Spec{Name: "test", Version: "0.0.1"},
		Components: map[string]*ir.Component{
			"http.server.api": {
				ID:   "http.server.api",
				Kind: ir.KindHTTPServer,
				HTTPServer: &ir.HTTPServerSpec{
					Framework: "hono",
					Port:      3000,
				},
			},
		},
	}

	gens, err := r.GeneratorsForIR(i)
	if err != nil {
		t.Fatalf("GeneratorsForIR() error = %v", err)
	}
	if len(gens) != 0 {
		t.Fatalf("GeneratorsForIR() len = %d, expected 0", len(gens))
	}
}

func TestPluginRegistry_RegisterDuplicate(t *testing.T) {
	r := NewPluginRegistry()
	plugin := GeneratorPlugin{
		Name: "dup",
		NewGenerator: func() Generator {
			return &mockGenerator{name: "dup"}
		},
	}

	if err := r.Register(plugin); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}
	if err := r.Register(plugin); err == nil {
		t.Fatal("expected duplicate plugin error")
	}
}
