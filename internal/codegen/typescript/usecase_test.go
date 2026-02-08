// Copyright 2026 OpenBoundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package typescript

import (
	"strings"
	"testing"

	"github.com/openboundary/openboundary/internal/codegen"
	"github.com/openboundary/openboundary/internal/ir"
	"github.com/openboundary/openboundary/internal/parser"
)

func TestNewUsecaseGenerator(t *testing.T) {
	// given/when
	g := NewUsecaseGenerator()

	// then
	if g == nil {
		t.Fatal("NewUsecaseGenerator() returned nil")
	}
}

func TestUsecaseGenerator_Name(t *testing.T) {
	// given
	g := NewUsecaseGenerator()

	// when
	name := g.Name()

	// then
	if name != "typescript-usecase" {
		t.Errorf("Name() = %q, want %q", name, "typescript-usecase")
	}
}

func TestUsecaseGenerator_Generate_TypesFile(t *testing.T) {
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
					BindsTo:    "http.server.api:POST:/users",
					Goal:       "Create a new user in the system",
					Actor:      "anonymous",
					Middleware: []string{},
					Preconditions: []string{
						"Email is not already registered",
					},
					AcceptanceCriteria: []string{
						"User record created",
						"Password hashed",
					},
					Postconditions: []string{
						"User exists in database",
					},
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
	g := NewUsecaseGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check types file
	typesFile, ok := output.Files["src/components/usecase-create-user.usecase.types.ts"]
	if !ok {
		t.Fatal("types file not found in output")
	}

	typesStr := string(typesFile.Content)

	// Types file should have DO NOT EDIT header
	if !strings.Contains(typesStr, "DO NOT EDIT") {
		t.Error("types file should contain DO NOT EDIT header")
	}

	// Types file should contain the function type alias
	if !strings.Contains(typesStr, "CreateUserUsecaseUsecaseFn") {
		t.Error("types file should contain function type alias")
	}

	// Types file should contain JSDoc
	if !strings.Contains(typesStr, "Create a new user in the system") {
		t.Error("types file should contain goal in JSDoc")
	}

	// Types file should contain preconditions
	if !strings.Contains(typesStr, "Email is not already registered") {
		t.Error("types file should contain preconditions")
	}

	// Types file should contain acceptance criteria
	if !strings.Contains(typesStr, "Password hashed") {
		t.Error("types file should contain acceptance criteria")
	}

	// Types file should use WriteAlways strategy
	if typesFile.Strategy != codegen.WriteAlways {
		t.Errorf("types file strategy = %v, expected WriteAlways", typesFile.Strategy)
	}
}

func TestUsecaseGenerator_Generate_ImplFile(t *testing.T) {
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
					BindsTo: "http.server.api:POST:/users",
					Goal:    "Create a new user in the system",
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
	g := NewUsecaseGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check impl file
	implFile, ok := output.Files["src/components/usecase-create-user.usecase.ts"]
	if !ok {
		t.Fatal("impl file not found in output")
	}

	implStr := string(implFile.Content)

	// Impl file should have the "will not be overwritten" header
	if !strings.Contains(implStr, "This file will not be overwritten") {
		t.Error("impl file should contain 'will not be overwritten' header")
	}

	// Impl file should import the type from types file
	if !strings.Contains(implStr, "from './usecase-create-user.usecase.types'") {
		t.Error("impl file should import from types file")
	}

	// Impl file should export a typed const
	if !strings.Contains(implStr, "export const createUserUsecase: CreateUserUsecaseUsecaseFn") {
		t.Error("impl file should contain typed const export")
	}

	// Impl file should use WriteOnce strategy
	if implFile.Strategy != codegen.WriteOnce {
		t.Errorf("impl file strategy = %v, expected WriteOnce", implFile.Strategy)
	}
}

func TestUsecaseGenerator_Generate_WithPathParams(t *testing.T) {
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
					BindsTo: "http.server.api:GET:/users/{id}",
					Goal:    "Get user by ID",
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
	g := NewUsecaseGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Path params should be in the types file
	content := string(output.Files["src/components/usecase-get-user.usecase.types.ts"].Content)

	// Check for path param in input type
	if !strings.Contains(content, "id: string") {
		t.Error("types file input should include path parameter 'id'")
	}
}

func TestUsecaseGenerator_Generate_WithAuthMiddleware(t *testing.T) {
	// given: usecase with auth middleware
	i := &ir.IR{
		Spec: &parser.Spec{Name: "test"},
		Components: map[string]*ir.Component{
			"http.server.api": {
				ID:   "http.server.api",
				Kind: ir.KindHTTPServer,
				HTTPServer: &ir.HTTPServerSpec{
					Framework:  "hono",
					Port:       3000,
					Middleware: []string{"middleware.authn"},
				},
			},
			"usecase.get-user": {
				ID:   "usecase.get-user",
				Kind: ir.KindUsecase,
				Usecase: &ir.UsecaseSpec{
					BindsTo:    "http.server.api:GET:/users/{id}",
					Goal:       "Get user by ID",
					Middleware: []string{"middleware.authn"},
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
	g := NewUsecaseGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Auth context should be in the types file
	content := string(output.Files["src/components/usecase-get-user.usecase.types.ts"].Content)

	// Check for auth in context type
	if !strings.Contains(content, "'auth'") {
		t.Error("types file context should include auth when middleware.authn is used")
	}
}

func TestUsecaseGenerator_Generate_IndexFile(t *testing.T) {
	// given: IR with multiple usecases
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
	g := NewUsecaseGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, ok := output.Files["src/components/usecases.ts"]
	if !ok {
		t.Fatal("index.ts not found in output")
	}

	contentStr := string(content.Content)

	// Check for exports
	if !strings.Contains(contentStr, "export { createUserUsecase }") {
		t.Error("index.ts should export createUserUsecase")
	}
	if !strings.Contains(contentStr, "export { getUserUsecase }") {
		t.Error("index.ts should export getUserUsecase")
	}
}

func TestUsecaseGenerator_Generate_NoUsecases(t *testing.T) {
	// given: IR with no usecases
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
		},
	}

	// when
	g := NewUsecaseGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Should still have index file
	if _, ok := output.Files["src/components/usecases.ts"]; !ok {
		t.Error("should generate index.ts even with no usecases")
	}

	// Should have only 1 file (index)
	if len(output.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(output.Files))
	}
}

func TestUsecaseGenerator_Generate_FileCount(t *testing.T) {
	// given: IR with 2 usecases
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
	g := NewUsecaseGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// 2 usecases * 2 files (types + impl) + 1 index = 5 files
	if len(output.Files) != 5 {
		t.Errorf("expected 5 files, got %d", len(output.Files))
	}
}
