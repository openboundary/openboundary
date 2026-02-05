// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package typescript

import (
	"strings"
	"testing"

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

func TestUsecaseGenerator_Generate_UsecaseFile(t *testing.T) {
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

	content, ok := output.Files["src/components/usecase-create-user.usecase.ts"]
	if !ok {
		t.Fatal("usecase file not found in output")
	}

	contentStr := string(content)

	// Check for function name
	if !strings.Contains(contentStr, "createUserUsecase") {
		t.Error("usecase file should contain createUserUsecase function")
	}

	// Check for JSDoc
	if !strings.Contains(contentStr, "Create a new user in the system") {
		t.Error("usecase file should contain goal in JSDoc")
	}

	// Check for preconditions
	if !strings.Contains(contentStr, "Email is not already registered") {
		t.Error("usecase file should contain preconditions")
	}

	// Check for acceptance criteria
	if !strings.Contains(contentStr, "Password hashed") {
		t.Error("usecase file should contain acceptance criteria")
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

	content := string(output.Files["src/components/usecase-get-user.usecase.ts"])

	// Check for path param in input type
	if !strings.Contains(content, "id: string") {
		t.Error("usecase input should include path parameter 'id'")
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

	content := string(output.Files["src/components/usecase-get-user.usecase.ts"])

	// Check for auth in context type
	if !strings.Contains(content, "'auth'") {
		t.Error("usecase context should include auth when middleware.authn is used")
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

	contentStr := string(content)

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
