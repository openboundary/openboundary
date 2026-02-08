// Copyright 2026 OpenBoundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package typescript

import (
	"testing"

	"github.com/openboundary/openboundary/internal/ir"
	"github.com/openboundary/openboundary/internal/parser"
)

func TestNewPluginRegistry(t *testing.T) {
	r, err := NewPluginRegistry()
	if err != nil {
		t.Fatalf("NewPluginRegistry() error = %v", err)
	}
	if r == nil {
		t.Fatal("NewPluginRegistry() returned nil")
	}
}

func TestNewPluginRegistry_GeneratorsForIR(t *testing.T) {
	r, err := NewPluginRegistry()
	if err != nil {
		t.Fatalf("NewPluginRegistry() error = %v", err)
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
	if len(gens) == 0 {
		t.Fatal("GeneratorsForIR() returned no generators")
	}
}
