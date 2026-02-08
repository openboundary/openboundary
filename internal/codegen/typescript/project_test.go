// Copyright 2026 OpenBoundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package typescript

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/openboundary/openboundary/internal/ir"
	"github.com/openboundary/openboundary/internal/parser"
)

func TestNewProjectGenerator(t *testing.T) {
	// given/when
	g := NewProjectGenerator()

	// then
	if g == nil {
		t.Fatal("NewProjectGenerator() returned nil")
	}
}

func TestProjectGenerator_Name(t *testing.T) {
	// given
	g := NewProjectGenerator()

	// when
	name := g.Name()

	// then
	if name != "typescript-project" {
		t.Errorf("Name() = %q, want %q", name, "typescript-project")
	}
}

func TestProjectGenerator_Generate_PackageJSON(t *testing.T) {
	// given
	i := &ir.IR{
		Spec: &parser.Spec{
			Name:        "test-api",
			Version:     "1.0.0",
			Description: "Test API",
		},
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

	// when
	g := NewProjectGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	pkgContent, ok := output.Files["package.json"]
	if !ok {
		t.Fatal("package.json not found in output")
	}

	var pkg PackageJSON
	if err := json.Unmarshal(pkgContent.Content, &pkg); err != nil {
		t.Fatalf("Failed to parse package.json: %v", err)
	}

	if pkg.Name != "test-api" {
		t.Errorf("package.json name = %q, want %q", pkg.Name, "test-api")
	}
	if pkg.Version != "1.0.0" {
		t.Errorf("package.json version = %q, want %q", pkg.Version, "1.0.0")
	}
	if pkg.Type != "module" {
		t.Errorf("package.json type = %q, want %q", pkg.Type, "module")
	}

	// Check required dependencies
	if _, ok := pkg.Dependencies["hono"]; !ok {
		t.Error("package.json missing hono dependency")
	}

	// Check scripts
	if _, ok := pkg.Scripts["test"]; !ok {
		t.Error("package.json missing test script")
	}
	if _, ok := pkg.Scripts["build"]; !ok {
		t.Error("package.json missing build script")
	}
}

func TestProjectGenerator_Generate_TSConfig(t *testing.T) {
	// given
	i := &ir.IR{
		Spec:       &parser.Spec{Name: "test"},
		Components: map[string]*ir.Component{},
	}

	// when
	g := NewProjectGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	tsContent, ok := output.Files["tsconfig.json"]
	if !ok {
		t.Fatal("tsconfig.json not found in output")
	}

	var config TSConfig
	if err := json.Unmarshal(tsContent.Content, &config); err != nil {
		t.Fatalf("Failed to parse tsconfig.json: %v", err)
	}

	if config.CompilerOptions.Target != "ES2022" {
		t.Errorf("tsconfig target = %q, want %q", config.CompilerOptions.Target, "ES2022")
	}
	if !config.CompilerOptions.Strict {
		t.Error("tsconfig strict should be true")
	}
	if config.CompilerOptions.OutDir != "./dist" {
		t.Errorf("tsconfig outDir = %q, want %q", config.CompilerOptions.OutDir, "./dist")
	}
}

func TestProjectGenerator_Generate_OrvalConfig(t *testing.T) {
	// given
	i := &ir.IR{
		Spec: &parser.Spec{Name: "test"},
		Components: map[string]*ir.Component{
			"http.server.api": {
				ID:   "http.server.api",
				Kind: ir.KindHTTPServer,
				HTTPServer: &ir.HTTPServerSpec{
					Framework: "hono",
					Port:      3000,
					OpenAPI:   "./src/components/http-server-api.openapi.yaml",
				},
			},
		},
	}

	// when
	g := NewProjectGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	orvalContent, ok := output.Files["orval.config.ts"]
	if !ok {
		t.Fatal("orval.config.ts not found in output")
	}

	content := string(orvalContent.Content)
	if !strings.Contains(content, "./src/components/http-server-api.openapi.yaml") {
		t.Error("orval.config.ts should contain the generated schema path")
	}
	if !strings.Contains(content, "defineConfig") {
		t.Error("orval.config.ts should use defineConfig")
	}
}

func TestProjectGenerator_Generate_DrizzleDependencies(t *testing.T) {
	// given: spec with postgres using drizzle
	i := &ir.IR{
		Spec: &parser.Spec{Name: "test"},
		Components: map[string]*ir.Component{
			"postgres.primary": {
				ID:   "postgres.primary",
				Kind: ir.KindPostgres,
				Postgres: &ir.PostgresSpec{
					Provider: "drizzle",
					Schema:   "./schema.ts",
				},
			},
		},
	}

	// when
	g := NewProjectGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	pkgContent := output.Files["package.json"]
	var pkg PackageJSON
	if err := json.Unmarshal(pkgContent.Content, &pkg); err != nil {
		t.Fatalf("Failed to parse package.json: %v", err)
	}

	if _, ok := pkg.Dependencies["drizzle-orm"]; !ok {
		t.Error("package.json should include drizzle-orm dependency")
	}
	if _, ok := pkg.DevDependencies["drizzle-kit"]; !ok {
		t.Error("package.json should include drizzle-kit devDependency")
	}
}

func TestProjectGenerator_Generate_BetterAuthDependencies(t *testing.T) {
	// given: spec with better-auth middleware
	i := &ir.IR{
		Spec: &parser.Spec{Name: "test"},
		Components: map[string]*ir.Component{
			"middleware.authn": {
				ID:   "middleware.authn",
				Kind: ir.KindMiddleware,
				Middleware: &ir.MiddlewareSpec{
					Provider: "better-auth",
					Config:   "./auth.config.ts",
				},
			},
		},
	}

	// when
	g := NewProjectGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	pkgContent := output.Files["package.json"]
	var pkg PackageJSON
	if err := json.Unmarshal(pkgContent.Content, &pkg); err != nil {
		t.Fatalf("Failed to parse package.json: %v", err)
	}

	if _, ok := pkg.Dependencies["better-auth"]; !ok {
		t.Error("package.json should include better-auth dependency")
	}
}

func TestProjectGenerator_Generate_CasbinDependencies(t *testing.T) {
	// given: spec with casbin middleware
	i := &ir.IR{
		Spec: &parser.Spec{Name: "test"},
		Components: map[string]*ir.Component{
			"middleware.authz": {
				ID:   "middleware.authz",
				Kind: ir.KindMiddleware,
				Middleware: &ir.MiddlewareSpec{
					Provider: "casbin",
					Model:    "./model.conf",
					Policy:   "./policy.csv",
				},
			},
		},
	}

	// when
	g := NewProjectGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	pkgContent := output.Files["package.json"]
	var pkg PackageJSON
	if err := json.Unmarshal(pkgContent.Content, &pkg); err != nil {
		t.Fatalf("Failed to parse package.json: %v", err)
	}

	if _, ok := pkg.Dependencies["casbin"]; !ok {
		t.Error("package.json should include casbin dependency")
	}
}

func TestProjectGenerator_Generate_GitIgnore(t *testing.T) {
	// given
	i := &ir.IR{
		Spec:       &parser.Spec{Name: "test"},
		Components: map[string]*ir.Component{},
	}

	// when
	g := NewProjectGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	gitignore, ok := output.Files[".gitignore"]
	if !ok {
		t.Fatal(".gitignore not found in output")
	}

	content := string(gitignore.Content)
	if !strings.Contains(content, "node_modules") {
		t.Error(".gitignore should contain node_modules")
	}
	if !strings.Contains(content, "dist") {
		t.Error(".gitignore should contain dist")
	}
}
