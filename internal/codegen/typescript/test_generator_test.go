// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package typescript

import (
	"strings"
	"testing"

	"github.com/openboundary/openboundary/internal/ir"
	"github.com/openboundary/openboundary/internal/parser"
)

func TestNewTestGenerator(t *testing.T) {
	// given/when
	g := NewTestGenerator()

	// then
	if g == nil {
		t.Fatal("NewTestGenerator() returned nil")
	}
}

func TestTestGenerator_Name(t *testing.T) {
	// given
	g := NewTestGenerator()

	// when
	name := g.Name()

	// then
	if name != "typescript-tests" {
		t.Errorf("Name() = %q, want %q", name, "typescript-tests")
	}
}

func TestTestGenerator_Generate_UsecaseTestFile(t *testing.T) {
	// given: IR with usecase
	i := &ir.IR{
		Spec: &parser.Spec{Name: "test"},
		Components: map[string]*ir.Component{
			"http.server.api": {
				ID:   "http.server.api",
				Kind: ir.KindHTTPServer,
				HTTPServer: &ir.HTTPServerSpec{
					Framework: "hono",
					Port:      3000,
				},
			},
			"usecase.create-user": {
				ID:   "usecase.create-user",
				Kind: ir.KindUsecase,
				Usecase: &ir.UsecaseSpec{
					Goal: "Create a new user",
					Binding: &ir.Binding{
						ServerID: "http.server.api",
						Method:   "POST",
						Path:     "/users",
					},
				},
			},
		},
	}

	// when
	g := NewTestGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, ok := output.Files["src/components/usecase-create-user.usecase.test.ts"]
	if !ok {
		t.Fatal("usecase test file not found in output")
	}

	contentStr := string(content.Content)

	// Check for vitest imports
	if !strings.Contains(contentStr, "import { describe, it, expect, vi, beforeEach } from 'vitest'") {
		t.Error("test file should import from vitest")
	}

	// Check for describe block
	if !strings.Contains(contentStr, "describe('createUserUsecase'") {
		t.Error("test file should have describe block for usecase")
	}

	// Check for BDD-style comments
	if !strings.Contains(contentStr, "// given") {
		t.Error("test file should have BDD-style given comments")
	}

	// Check for basic tests
	if !strings.Contains(contentStr, "should be a function") {
		t.Error("test file should test that usecase is a function")
	}
	if !strings.Contains(contentStr, "should return a promise") {
		t.Error("test file should test that usecase returns a promise")
	}
}

func TestTestGenerator_Generate_UsecaseWithPathParams(t *testing.T) {
	// given: usecase with path parameters
	i := &ir.IR{
		Spec: &parser.Spec{Name: "test"},
		Components: map[string]*ir.Component{
			"http.server.api": {
				ID:   "http.server.api",
				Kind: ir.KindHTTPServer,
				HTTPServer: &ir.HTTPServerSpec{
					Framework: "hono",
					Port:      3000,
				},
			},
			"usecase.get-user": {
				ID:   "usecase.get-user",
				Kind: ir.KindUsecase,
				Usecase: &ir.UsecaseSpec{
					Goal: "Get user by ID",
					Binding: &ir.Binding{
						ServerID: "http.server.api",
						Method:   "GET",
						Path:     "/users/{id}",
					},
				},
			},
		},
	}

	// when
	g := NewTestGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(output.Files["src/components/usecase-get-user.usecase.test.ts"].Content)

	// Should have test for path params
	if !strings.Contains(content, "should accept path parameters in input") {
		t.Error("test file should have test for path parameters")
	}
	if !strings.Contains(content, "id: 'test-id'") {
		t.Error("test file should include test value for id param")
	}
}

func TestTestGenerator_Generate_UsecaseWithAuthMiddleware(t *testing.T) {
	// given: usecase with auth middleware
	i := &ir.IR{
		Spec: &parser.Spec{Name: "test"},
		Components: map[string]*ir.Component{
			"http.server.api": {
				ID:   "http.server.api",
				Kind: ir.KindHTTPServer,
				HTTPServer: &ir.HTTPServerSpec{
					Framework: "hono",
					Port:      3000,
				},
			},
			"usecase.get-profile": {
				ID:   "usecase.get-profile",
				Kind: ir.KindUsecase,
				Usecase: &ir.UsecaseSpec{
					Goal:       "Get current user profile",
					Middleware: []string{"middleware.authn"},
					Binding: &ir.Binding{
						ServerID: "http.server.api",
						Method:   "GET",
						Path:     "/profile",
					},
				},
			},
		},
	}

	// when
	g := NewTestGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(output.Files["src/components/usecase-get-profile.usecase.test.ts"].Content)

	// Should have auth context test
	if !strings.Contains(content, "should have access to auth context") {
		t.Error("test file should have test for auth context")
	}
}

func TestTestGenerator_Generate_MiddlewareTestFile(t *testing.T) {
	// given: IR with middleware
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
	g := NewTestGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, ok := output.Files["src/components/middleware-authn.middleware.test.ts"]
	if !ok {
		t.Fatal("middleware test file not found in output")
	}

	contentStr := string(content.Content)

	// Check for describe block
	if !strings.Contains(contentStr, "describe('middlewareAuthnMiddleware'") {
		t.Error("test file should have describe block for middleware")
	}

	// Check for middleware function test
	if !strings.Contains(contentStr, "should be a middleware function") {
		t.Error("test file should test middleware is a function")
	}

	// Check for better-auth specific test
	if !strings.Contains(contentStr, "should set auth in context when session is valid") {
		t.Error("test file should have better-auth specific test")
	}

	// Should NOT call middleware as factory (no parentheses after function name)
	if strings.Contains(contentStr, "middlewareAuthnMiddleware()") {
		t.Error("test file should not call middleware as factory - middleware is already a function")
	}
}

func TestTestGenerator_Generate_CasbinMiddlewareTest(t *testing.T) {
	// given: IR with casbin middleware
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
	g := NewTestGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(output.Files["src/components/middleware-authz.middleware.test.ts"].Content)

	// Check for casbin specific test
	if !strings.Contains(content, "should check authorization using enforcer") {
		t.Error("test file should have casbin-specific test")
	}
}

func TestTestGenerator_Generate_ServerTestFile(t *testing.T) {
	// given: IR with server and bound usecases
	i := &ir.IR{
		Spec: &parser.Spec{Name: "test"},
		Components: map[string]*ir.Component{
			"http.server.api": {
				ID:   "http.server.api",
				Kind: ir.KindHTTPServer,
				HTTPServer: &ir.HTTPServerSpec{
					Framework: "hono",
					Port:      3000,
				},
			},
			"usecase.create-user": {
				ID:   "usecase.create-user",
				Kind: ir.KindUsecase,
				Usecase: &ir.UsecaseSpec{
					Goal: "Create user",
					Binding: &ir.Binding{
						ServerID: "http.server.api",
						Method:   "POST",
						Path:     "/users",
					},
				},
			},
			"usecase.get-user": {
				ID:   "usecase.get-user",
				Kind: ir.KindUsecase,
				Usecase: &ir.UsecaseSpec{
					Goal: "Get user",
					Binding: &ir.Binding{
						ServerID: "http.server.api",
						Method:   "GET",
						Path:     "/users/{id}",
					},
				},
			},
		},
	}

	// when
	g := NewTestGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, ok := output.Files["src/components/http-server-api.server.test.ts"]
	if !ok {
		t.Fatal("server test file not found in output")
	}

	contentStr := string(content.Content)

	// Check for describe block
	if !strings.Contains(contentStr, "describe('createHttpServerApiApp'") {
		t.Error("test file should have describe block for server")
	}

	// Check for app creation test
	if !strings.Contains(contentStr, "should create a Hono app instance") {
		t.Error("test file should test app creation")
	}

	// Check for route tests
	if !strings.Contains(contentStr, "should have POST /users route") {
		t.Error("test file should have POST /users route test")
	}
	if !strings.Contains(contentStr, "should have GET /users/:id route") {
		t.Error("test file should have GET /users/:id route test")
	}
}

func TestTestGenerator_Generate_TestSetupFile(t *testing.T) {
	// given: any IR
	i := &ir.IR{
		Spec:       &parser.Spec{Name: "test"},
		Components: map[string]*ir.Component{},
	}

	// when
	g := NewTestGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, ok := output.Files["src/test/setup.ts"]
	if !ok {
		t.Fatal("test setup file not found in output")
	}

	contentStr := string(content.Content)

	// Check for mock context creator
	if !strings.Contains(contentStr, "createMockContext") {
		t.Error("setup file should have createMockContext function")
	}

	// Check for mock Hono context creator
	if !strings.Contains(contentStr, "createMockHonoContext") {
		t.Error("setup file should have createMockHonoContext function")
	}

	// Check for db mock
	if !strings.Contains(contentStr, "db:") {
		t.Error("setup file should mock db")
	}

	// Check for auth mock
	if !strings.Contains(contentStr, "auth:") {
		t.Error("setup file should mock auth")
	}

	// Check for enforcer mock
	if !strings.Contains(contentStr, "enforcer:") {
		t.Error("setup file should mock enforcer")
	}
}

func TestTestGenerator_Generate_NoComponents(t *testing.T) {
	// given: IR with no testable components
	i := &ir.IR{
		Spec:       &parser.Spec{Name: "test"},
		Components: map[string]*ir.Component{},
	}

	// when
	g := NewTestGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Should still have setup file
	if _, ok := output.Files["src/test/setup.ts"]; !ok {
		t.Error("should generate test setup file even with no components")
	}

	// Should have only 1 file (setup)
	if len(output.Files) != 1 {
		t.Errorf("expected 1 file (setup), got %d", len(output.Files))
	}
}

func TestTestGenerator_Generate_ServerWithAuthMiddleware(t *testing.T) {
	// given: server with auth middleware
	i := &ir.IR{
		Spec: &parser.Spec{Name: "test"},
		Components: map[string]*ir.Component{
			"http.server.api": {
				ID:   "http.server.api",
				Kind: ir.KindHTTPServer,
				HTTPServer: &ir.HTTPServerSpec{
					Framework:  "hono",
					Port:       3000,
					Middleware: []string{"middleware.authn", "middleware.authz"},
				},
			},
		},
	}

	// when
	g := NewTestGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(output.Files["src/components/http-server-api.server.test.ts"].Content)

	// Should have auth mock in createMockDeps
	if !strings.Contains(content, "auth:") {
		t.Error("server test should mock auth for authn middleware")
	}

	// Should have enforcer mock
	if !strings.Contains(content, "enforcer:") {
		t.Error("server test should mock enforcer for authz middleware")
	}
}
