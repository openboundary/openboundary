// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package typescript

import (
	"strings"
	"testing"

	"github.com/openboundary/openboundary/internal/ir"
	"github.com/openboundary/openboundary/internal/parser"
)

func TestNewContextGenerator(t *testing.T) {
	// given/when
	g := NewContextGenerator()

	// then
	if g == nil {
		t.Fatal("NewContextGenerator() returned nil")
	}
}

func TestContextGenerator_Name(t *testing.T) {
	// given
	g := NewContextGenerator()

	// when
	name := g.Name()

	// then
	if name != "typescript-context" {
		t.Errorf("Name() = %q, want %q", name, "typescript-context")
	}
}

func TestContextGenerator_Generate_WithPostgresDependency(t *testing.T) {
	// given: server with postgres dependency
	postgres := &ir.Component{
		ID:   "postgres.primary",
		Kind: ir.KindPostgres,
		Postgres: &ir.PostgresSpec{
			Provider: "drizzle",
			Schema:   "./schema.ts",
		},
	}

	server := &ir.Component{
		ID:   "http.server.api",
		Kind: ir.KindHTTPServer,
		HTTPServer: &ir.HTTPServerSpec{
			Framework: "hono",
			Port:      3000,
		},
		Dependencies: []*ir.Component{postgres},
	}

	i := &ir.IR{
		Spec: &parser.Spec{Name: "test"},
		Components: map[string]*ir.Component{
			"http.server.api":  server,
			"postgres.primary": postgres,
		},
	}

	// when
	g := NewContextGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, ok := output.Files["src/components/http-server-api.context.ts"]
	if !ok {
		t.Fatal("context file not found in output")
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "ServerContext") {
		t.Error("context file should contain ServerContext interface")
	}
	if !strings.Contains(contentStr, "DrizzleClient") {
		t.Error("context file should reference DrizzleClient type")
	}
	if !strings.Contains(contentStr, "db:") {
		t.Error("context file should have db field")
	}
}

func TestContextGenerator_Generate_WithBetterAuthMiddleware(t *testing.T) {
	// given: server with better-auth middleware
	mw := &ir.Component{
		ID:   "middleware.authn",
		Kind: ir.KindMiddleware,
		Middleware: &ir.MiddlewareSpec{
			Provider: "better-auth",
			Config:   "./auth.config.ts",
		},
	}

	server := &ir.Component{
		ID:   "http.server.api",
		Kind: ir.KindHTTPServer,
		HTTPServer: &ir.HTTPServerSpec{
			Framework:  "hono",
			Port:       3000,
			Middleware: []string{"middleware.authn"},
		},
	}

	i := &ir.IR{
		Spec: &parser.Spec{Name: "test"},
		Components: map[string]*ir.Component{
			"http.server.api":  server,
			"middleware.authn": mw,
		},
	}

	// when
	g := NewContextGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(output.Files["src/components/http-server-api.context.ts"])
	if !strings.Contains(content, "auth?:") {
		t.Error("context file should have auth field for better-auth")
	}
	if !strings.Contains(content, "Session") {
		t.Error("context file should reference session property")
	}
}

func TestContextGenerator_Generate_WithCasbinMiddleware(t *testing.T) {
	// given: server with casbin middleware
	mw := &ir.Component{
		ID:   "middleware.authz",
		Kind: ir.KindMiddleware,
		Middleware: &ir.MiddlewareSpec{
			Provider: "casbin",
			Model:    "./model.conf",
			Policy:   "./policy.csv",
		},
	}

	server := &ir.Component{
		ID:   "http.server.api",
		Kind: ir.KindHTTPServer,
		HTTPServer: &ir.HTTPServerSpec{
			Framework:  "hono",
			Port:       3000,
			Middleware: []string{"middleware.authz"},
		},
	}

	i := &ir.IR{
		Spec: &parser.Spec{Name: "test"},
		Components: map[string]*ir.Component{
			"http.server.api":  server,
			"middleware.authz": mw,
		},
	}

	// when
	g := NewContextGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(output.Files["src/components/http-server-api.context.ts"])
	if !strings.Contains(content, "enforcer?:") {
		t.Error("context file should have enforcer field for casbin")
	}
	if !strings.Contains(content, "Enforcer") {
		t.Error("context file should reference Enforcer type")
	}
}

func TestContextGenerator_Generate_ContextWithHelper(t *testing.T) {
	// given: any server
	server := &ir.Component{
		ID:   "http.server.api",
		Kind: ir.KindHTTPServer,
		HTTPServer: &ir.HTTPServerSpec{
			Framework: "hono",
			Port:      3000,
		},
	}

	i := &ir.IR{
		Spec: &parser.Spec{Name: "test"},
		Components: map[string]*ir.Component{
			"http.server.api": server,
		},
	}

	// when
	g := NewContextGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(output.Files["src/components/http-server-api.context.ts"])
	if !strings.Contains(content, "ContextWith") {
		t.Error("context file should contain ContextWith helper type")
	}
	if !strings.Contains(content, "Pick<ServerContext") {
		t.Error("ContextWith should use Pick utility type")
	}
}

func TestContextGenerator_Generate_NoHTTPServers(t *testing.T) {
	// given: IR with no http.server components
	i := &ir.IR{
		Spec: &parser.Spec{Name: "test"},
		Components: map[string]*ir.Component{
			"postgres.primary": {
				ID:   "postgres.primary",
				Kind: ir.KindPostgres,
				Postgres: &ir.PostgresSpec{
					Provider: "drizzle",
				},
			},
		},
	}

	// when
	g := NewContextGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(output.Files) != 0 {
		t.Errorf("expected no files for IR without http.server, got %d", len(output.Files))
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"http.server.api", "http-server-api"},
		{"middleware.authn", "middleware-authn"},
		{"postgres.primary", "postgres-primary"},
		{"simple", "simple"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// when
			got := sanitizeFilename(tt.input)

			// then
			if got != tt.want {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
