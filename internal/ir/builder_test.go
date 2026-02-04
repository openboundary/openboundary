package ir

import (
	"strings"
	"testing"

	"github.com/stack-bound/stack-bound/internal/parser"
)

func TestNewBuilder(t *testing.T) {
	b := NewBuilder()
	if b == nil {
		t.Fatal("NewBuilder() returned nil")
	}
}

func TestBuilder_Build(t *testing.T) {
	tests := []struct {
		name           string
		spec           *parser.Spec
		wantComponents int
		wantEdges      int
		wantErrors     int
	}{
		{
			name: "empty spec",
			spec: &parser.Spec{
				Components: []parser.Component{},
			},
			wantComponents: 0,
			wantEdges:      0,
			wantErrors:     0,
		},
		{
			name: "single http.server",
			spec: &parser.Spec{
				Components: []parser.Component{
					{
						ID:   "http.server.api",
						Kind: "http.server",
						Spec: map[string]interface{}{
							"framework": "hono",
							"port":      3000,
						},
					},
				},
			},
			wantComponents: 1,
			wantEdges:      0,
			wantErrors:     0,
		},
		{
			name: "http.server with middleware dependency",
			spec: &parser.Spec{
				Components: []parser.Component{
					{
						ID:   "middleware.auth",
						Kind: "middleware",
						Spec: map[string]interface{}{
							"provider": "better-auth",
							"config":   "./auth.ts",
						},
					},
					{
						ID:   "http.server.api",
						Kind: "http.server",
						Spec: map[string]interface{}{
							"framework":  "hono",
							"port":       3000,
							"middleware": []interface{}{"middleware.auth"},
						},
					},
				},
			},
			wantComponents: 2,
			wantEdges:      1,
			wantErrors:     0,
		},
		{
			name: "unknown kind",
			spec: &parser.Spec{
				Components: []parser.Component{
					{
						ID:   "unknown.comp",
						Kind: "unknown",
						Spec: map[string]interface{}{},
					},
				},
			},
			wantComponents: 0,
			wantEdges:      0,
			wantErrors:     1,
		},
		{
			name: "unresolved reference",
			spec: &parser.Spec{
				Components: []parser.Component{
					{
						ID:   "http.server.api",
						Kind: "http.server",
						Spec: map[string]interface{}{
							"framework":  "hono",
							"port":       3000,
							"middleware": []interface{}{"nonexistent.middleware"},
						},
					},
				},
			},
			wantComponents: 1,
			wantEdges:      0,
			wantErrors:     1,
		},
		{
			name: "usecase with binding",
			spec: &parser.Spec{
				Components: []parser.Component{
					{
						ID:   "http.server.api",
						Kind: "http.server",
						Spec: map[string]interface{}{
							"framework": "hono",
							"port":      3000,
						},
					},
					{
						ID:   "usecase.create-user",
						Kind: "usecase",
						Spec: map[string]interface{}{
							"binds_to": "http.server.api:POST:/users",
							"goal":     "Create a user",
						},
					},
				},
			},
			wantComponents: 2,
			wantEdges:      1,
			wantErrors:     0,
		},
		{
			name: "all component types",
			spec: &parser.Spec{
				Components: []parser.Component{
					{
						ID:   "http.server.api",
						Kind: "http.server",
						Spec: map[string]interface{}{
							"framework":  "hono",
							"port":       float64(3000),
							"depends_on": []interface{}{"postgres.primary"},
						},
					},
					{
						ID:   "middleware.auth",
						Kind: "middleware",
						Spec: map[string]interface{}{
							"provider": "better-auth",
							"config":   "./auth.ts",
						},
					},
					{
						ID:   "postgres.primary",
						Kind: "postgres",
						Spec: map[string]interface{}{
							"provider": "drizzle",
							"schema":   "./schema.ts",
						},
					},
					{
						ID:   "usecase.create-user",
						Kind: "usecase",
						Spec: map[string]interface{}{
							"binds_to":            "http.server.api:POST:/users",
							"goal":                "Create a user",
							"middleware":          []interface{}{"middleware.auth"},
							"actor":               "User",
							"preconditions":       []interface{}{"User is not logged in"},
							"acceptance_criteria": []interface{}{"User is created"},
							"postconditions":      []interface{}{"User exists in database"},
						},
					},
				},
			},
			wantComponents: 4,
			wantEdges:      3, // server->postgres, usecase->server, usecase->middleware
			wantErrors:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder()
			ir, errs := b.Build(tt.spec)

			if len(errs) != tt.wantErrors {
				t.Errorf("Build() returned %d errors, expected %d", len(errs), tt.wantErrors)
				for _, e := range errs {
					t.Logf("  error: %v", e)
				}
			}

			if len(ir.Components) != tt.wantComponents {
				t.Errorf("Build() created %d components, expected %d", len(ir.Components), tt.wantComponents)
			}

			if len(ir.Edges) != tt.wantEdges {
				t.Errorf("Build() created %d edges, expected %d", len(ir.Edges), tt.wantEdges)
			}
		})
	}
}

func TestBuilder_Build_HTTPServerSpec(t *testing.T) {
	spec := &parser.Spec{
		Components: []parser.Component{
			{
				ID:   "http.server.api",
				Kind: "http.server",
				Spec: map[string]interface{}{
					"framework": "hono",
					"port":      float64(3000),
					"openapi":   "./openapi.yaml",
				},
			},
		},
	}

	b := NewBuilder()
	ir, _ := b.Build(spec)

	comp := ir.Components["http.server.api"]
	if comp == nil {
		t.Fatal("component not found")
	}
	if comp.HTTPServer == nil {
		t.Fatal("HTTPServer spec is nil")
	}
	if comp.HTTPServer.Framework != "hono" {
		t.Errorf("Framework = %q, expected %q", comp.HTTPServer.Framework, "hono")
	}
	if comp.HTTPServer.Port != 3000 {
		t.Errorf("Port = %d, expected %d", comp.HTTPServer.Port, 3000)
	}
	if comp.HTTPServer.OpenAPI != "./openapi.yaml" {
		t.Errorf("OpenAPI = %q, expected %q", comp.HTTPServer.OpenAPI, "./openapi.yaml")
	}
}

func TestBuilder_Build_MiddlewareSpec(t *testing.T) {
	spec := &parser.Spec{
		Components: []parser.Component{
			{
				ID:   "middleware.authz",
				Kind: "middleware",
				Spec: map[string]interface{}{
					"provider": "casbin",
					"model":    "./model.conf",
					"policy":   "./policy.csv",
				},
			},
		},
	}

	b := NewBuilder()
	ir, _ := b.Build(spec)

	comp := ir.Components["middleware.authz"]
	if comp == nil {
		t.Fatal("component not found")
	}
	if comp.Middleware == nil {
		t.Fatal("Middleware spec is nil")
	}
	if comp.Middleware.Provider != "casbin" {
		t.Errorf("Provider = %q, expected %q", comp.Middleware.Provider, "casbin")
	}
	if comp.Middleware.Model != "./model.conf" {
		t.Errorf("Model = %q, expected %q", comp.Middleware.Model, "./model.conf")
	}
	if comp.Middleware.Policy != "./policy.csv" {
		t.Errorf("Policy = %q, expected %q", comp.Middleware.Policy, "./policy.csv")
	}
}

func TestBuilder_Build_PostgresSpec(t *testing.T) {
	spec := &parser.Spec{
		Components: []parser.Component{
			{
				ID:   "postgres.primary",
				Kind: "postgres",
				Spec: map[string]interface{}{
					"provider": "drizzle",
					"schema":   "./schema.ts",
				},
			},
		},
	}

	b := NewBuilder()
	ir, _ := b.Build(spec)

	comp := ir.Components["postgres.primary"]
	if comp == nil {
		t.Fatal("component not found")
	}
	if comp.Postgres == nil {
		t.Fatal("Postgres spec is nil")
	}
	if comp.Postgres.Provider != "drizzle" {
		t.Errorf("Provider = %q, expected %q", comp.Postgres.Provider, "drizzle")
	}
	if comp.Postgres.Schema != "./schema.ts" {
		t.Errorf("Schema = %q, expected %q", comp.Postgres.Schema, "./schema.ts")
	}
}

func TestBuilder_Build_UsecaseSpec(t *testing.T) {
	spec := &parser.Spec{
		Components: []parser.Component{
			{
				ID:   "http.server.api",
				Kind: "http.server",
				Spec: map[string]interface{}{
					"framework": "hono",
					"port":      3000,
				},
			},
			{
				ID:   "usecase.test",
				Kind: "usecase",
				Spec: map[string]interface{}{
					"binds_to":            "http.server.api:GET:/test",
					"goal":                "Test goal",
					"actor":               "User",
					"preconditions":       []interface{}{"pre1"},
					"acceptance_criteria": []interface{}{"ac1", "ac2"},
					"postconditions":      []interface{}{"post1"},
				},
			},
		},
	}

	b := NewBuilder()
	ir, _ := b.Build(spec)

	comp := ir.Components["usecase.test"]
	if comp == nil {
		t.Fatal("component not found")
	}
	if comp.Usecase == nil {
		t.Fatal("Usecase spec is nil")
	}
	if comp.Usecase.BindsTo != "http.server.api:GET:/test" {
		t.Errorf("BindsTo = %q", comp.Usecase.BindsTo)
	}
	if comp.Usecase.Goal != "Test goal" {
		t.Errorf("Goal = %q", comp.Usecase.Goal)
	}
	if comp.Usecase.Actor != "User" {
		t.Errorf("Actor = %q", comp.Usecase.Actor)
	}
	if len(comp.Usecase.Preconditions) != 1 {
		t.Errorf("Preconditions = %v", comp.Usecase.Preconditions)
	}
	if len(comp.Usecase.AcceptanceCriteria) != 2 {
		t.Errorf("AcceptanceCriteria = %v", comp.Usecase.AcceptanceCriteria)
	}
	if len(comp.Usecase.Postconditions) != 1 {
		t.Errorf("Postconditions = %v", comp.Usecase.Postconditions)
	}
}

func TestExtractServerFromBinding(t *testing.T) {
	tests := []struct {
		bindsTo  string
		expected string
	}{
		{"http.server.api:POST:/users", "http.server.api"},
		{"server:GET:/", "server"},
		{"no-colon", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.bindsTo, func(t *testing.T) {
			got := extractServerFromBinding(tt.bindsTo)
			if got != tt.expected {
				t.Errorf("extractServerFromBinding(%q) = %q, expected %q", tt.bindsTo, got, tt.expected)
			}
		})
	}
}

func TestToStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []string
	}{
		{"empty", []interface{}{}, []string{}},
		{"strings", []interface{}{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"mixed", []interface{}{"a", 123, "b"}, []string{"a", "b"}},
		{"nil", nil, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toStringSlice(tt.input)
			if len(got) != len(tt.expected) {
				t.Errorf("toStringSlice() len = %d, expected %d", len(got), len(tt.expected))
				return
			}
			for i := range tt.expected {
				if got[i] != tt.expected[i] {
					t.Errorf("toStringSlice()[%d] = %q, expected %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestBuilder_Build_MiddlewareDependsOn(t *testing.T) {
	spec := &parser.Spec{
		Components: []parser.Component{
			{
				ID:   "middleware.authn",
				Kind: "middleware",
				Spec: map[string]interface{}{
					"provider": "better-auth",
					"config":   "./auth.ts",
				},
			},
			{
				ID:   "middleware.authz",
				Kind: "middleware",
				Spec: map[string]interface{}{
					"provider":   "casbin",
					"model":      "./model.conf",
					"policy":     "./policy.csv",
					"depends_on": []interface{}{"middleware.authn"},
				},
			},
		},
	}

	b := NewBuilder()
	ir, errs := b.Build(spec)

	if len(errs) != 0 {
		t.Errorf("Build() returned errors: %v", errs)
	}

	if len(ir.Edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(ir.Edges))
	}

	authz := ir.Components["middleware.authz"]
	if len(authz.Dependencies) != 1 {
		t.Errorf("authz has %d dependencies, expected 1", len(authz.Dependencies))
	}
}

func TestBuilder_Build_DuplicateComponentID(t *testing.T) {
	spec := &parser.Spec{
		Components: []parser.Component{
			{
				ID:   "comp.dup",
				Kind: "http.server",
				Spec: map[string]interface{}{
					"framework": "hono",
					"port":      3000,
				},
			},
			{
				ID:   "comp.dup",
				Kind: "http.server",
				Spec: map[string]interface{}{
					"framework": "hono",
					"port":      3001,
				},
			},
		},
	}

	b := NewBuilder()
	_, errs := b.Build(spec)

	if len(errs) == 0 {
		t.Error("Build() expected error for duplicate component ID")
	}
}

func TestBuilder_Build_UsecaseNoBindsTo(t *testing.T) {
	spec := &parser.Spec{
		Components: []parser.Component{
			{
				ID:   "usecase.test",
				Kind: "usecase",
				Spec: map[string]interface{}{
					"goal": "Test",
					// No binds_to - should still build without error, validation will catch it
				},
			},
		},
	}

	b := NewBuilder()
	ir, errs := b.Build(spec)

	// Build should succeed, validation should fail
	if len(errs) != 0 {
		t.Errorf("Build() should not error without binds_to, got: %v", errs)
	}

	comp := ir.Components["usecase.test"]
	if comp.Usecase.BindsTo != "" {
		t.Errorf("BindsTo should be empty")
	}
}

func TestBuilder_Build_UsecaseMiddlewareRef(t *testing.T) {
	spec := &parser.Spec{
		Components: []parser.Component{
			{
				ID:   "http.server.api",
				Kind: "http.server",
				Spec: map[string]interface{}{
					"framework": "hono",
					"port":      3000,
				},
			},
			{
				ID:   "middleware.auth",
				Kind: "middleware",
				Spec: map[string]interface{}{
					"provider": "better-auth",
					"config":   "./auth.ts",
				},
			},
			{
				ID:   "usecase.test",
				Kind: "usecase",
				Spec: map[string]interface{}{
					"binds_to":   "http.server.api:GET:/test",
					"goal":       "Test",
					"middleware": []interface{}{"middleware.auth"},
				},
			},
		},
	}

	b := NewBuilder()
	ir, errs := b.Build(spec)

	if len(errs) != 0 {
		t.Errorf("Build() errors: %v", errs)
	}

	// Should have 2 edges: usecase->server (binding), usecase->middleware
	if len(ir.Edges) != 2 {
		t.Errorf("Expected 2 edges, got %d", len(ir.Edges))
	}
}

func TestBuilder_Build_UnresolvedHTTPServerDependsOn(t *testing.T) {
	spec := &parser.Spec{
		Components: []parser.Component{
			{
				ID:   "http.server.api",
				Kind: "http.server",
				Spec: map[string]interface{}{
					"framework":  "hono",
					"port":       3000,
					"depends_on": []interface{}{"nonexistent.comp"},
				},
			},
		},
	}

	b := NewBuilder()
	_, errs := b.Build(spec)

	if len(errs) != 1 {
		t.Errorf("Build() expected 1 error for unresolved depends_on, got %d", len(errs))
	}
}

func TestBuilder_Build_UnresolvedMiddlewareDependsOn(t *testing.T) {
	spec := &parser.Spec{
		Components: []parser.Component{
			{
				ID:   "middleware.auth",
				Kind: "middleware",
				Spec: map[string]interface{}{
					"provider":   "better-auth",
					"config":     "./auth.ts",
					"depends_on": []interface{}{"nonexistent.middleware"},
				},
			},
		},
	}

	b := NewBuilder()
	_, errs := b.Build(spec)

	if len(errs) != 1 {
		t.Errorf("Build() expected 1 error for unresolved middleware depends_on, got %d", len(errs))
	}
}

func TestBuilder_Build_UnresolvedUsecaseBindsTo(t *testing.T) {
	spec := &parser.Spec{
		Components: []parser.Component{
			{
				ID:   "usecase.test",
				Kind: "usecase",
				Spec: map[string]interface{}{
					"binds_to": "nonexistent.server:GET:/test",
					"goal":     "Test",
				},
			},
		},
	}

	b := NewBuilder()
	_, errs := b.Build(spec)

	// Expect errors from both reference resolution and usecase linking phases
	if len(errs) < 1 {
		t.Errorf("Build() expected errors for unresolved binds_to, got %d", len(errs))
	}

	// Verify at least one error mentions the unresolved server
	hasServerError := false
	for _, err := range errs {
		if err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "nonexistent.server") || strings.Contains(errStr, "unresolved") {
				hasServerError = true
				break
			}
		}
	}
	if !hasServerError {
		t.Errorf("Build() errors should mention unresolved server reference")
	}
}

func TestBuilder_Build_UnresolvedUsecaseMiddleware(t *testing.T) {
	spec := &parser.Spec{
		Components: []parser.Component{
			{
				ID:   "http.server.api",
				Kind: "http.server",
				Spec: map[string]interface{}{
					"framework": "hono",
					"port":      3000,
				},
			},
			{
				ID:   "usecase.test",
				Kind: "usecase",
				Spec: map[string]interface{}{
					"binds_to":   "http.server.api:GET:/test",
					"goal":       "Test",
					"middleware": []interface{}{"nonexistent.middleware"},
				},
			},
		},
	}

	b := NewBuilder()
	_, errs := b.Build(spec)

	if len(errs) != 1 {
		t.Errorf("Build() expected 1 error for unresolved usecase middleware, got %d", len(errs))
	}
}

func TestBuilder_Build_PortAsInt(t *testing.T) {
	spec := &parser.Spec{
		Components: []parser.Component{
			{
				ID:   "http.server.api",
				Kind: "http.server",
				Spec: map[string]interface{}{
					"framework": "hono",
					"port":      3000, // int, not float64
				},
			},
		},
	}

	b := NewBuilder()
	ir, _ := b.Build(spec)

	comp := ir.Components["http.server.api"]
	if comp.HTTPServer.Port != 3000 {
		t.Errorf("Port = %d, expected 3000", comp.HTTPServer.Port)
	}
}

func TestBuilder_WithBaseDir(t *testing.T) {
	// given
	b := NewBuilder()

	// when
	result := b.WithBaseDir("/some/path")

	// then
	if result != b {
		t.Error("WithBaseDir should return the same builder for chaining")
	}
	if b.baseDir != "/some/path" {
		t.Errorf("baseDir = %q, expected %q", b.baseDir, "/some/path")
	}
}

func TestBuilder_Build_UsecaseBinding(t *testing.T) {
	// given: a spec with usecase bound to server
	spec := &parser.Spec{
		Components: []parser.Component{
			{
				ID:   "http.server.api",
				Kind: "http.server",
				Spec: map[string]interface{}{
					"framework": "hono",
					"port":      3000,
				},
			},
			{
				ID:   "usecase.create-user",
				Kind: "usecase",
				Spec: map[string]interface{}{
					"binds_to": "http.server.api:POST:/users",
					"goal":     "Create a user",
				},
			},
		},
	}

	// when
	b := NewBuilder()
	ir, errs := b.Build(spec)

	// then
	if len(errs) > 0 {
		t.Fatalf("Build() unexpected errors: %v", errs)
	}

	usecase := ir.Components["usecase.create-user"]
	if usecase.Usecase.Binding == nil {
		t.Fatal("Binding should not be nil")
	}
	if usecase.Usecase.Binding.ServerID != "http.server.api" {
		t.Errorf("Binding.ServerID = %q, expected %q", usecase.Usecase.Binding.ServerID, "http.server.api")
	}
	if usecase.Usecase.Binding.Method != "POST" {
		t.Errorf("Binding.Method = %q, expected %q", usecase.Usecase.Binding.Method, "POST")
	}
	if usecase.Usecase.Binding.Path != "/users" {
		t.Errorf("Binding.Path = %q, expected %q", usecase.Usecase.Binding.Path, "/users")
	}
}

func TestBuilder_Build_UsecaseBindingWithPathParams(t *testing.T) {
	// given: a spec with usecase bound to path with parameters
	spec := &parser.Spec{
		Components: []parser.Component{
			{
				ID:   "http.server.api",
				Kind: "http.server",
				Spec: map[string]interface{}{
					"framework": "hono",
					"port":      3000,
				},
			},
			{
				ID:   "usecase.get-user",
				Kind: "usecase",
				Spec: map[string]interface{}{
					"binds_to": "http.server.api:GET:/users/{id}",
					"goal":     "Get a user",
				},
			},
		},
	}

	// when
	b := NewBuilder()
	ir, errs := b.Build(spec)

	// then
	if len(errs) > 0 {
		t.Fatalf("Build() unexpected errors: %v", errs)
	}

	usecase := ir.Components["usecase.get-user"]
	if usecase.Usecase.Binding.Path != "/users/{id}" {
		t.Errorf("Binding.Path = %q, expected %q", usecase.Usecase.Binding.Path, "/users/{id}")
	}
}

func TestBuilder_Build_InvalidBindsToFormat(t *testing.T) {
	tests := []struct {
		name    string
		bindsTo string
	}{
		{
			name:    "missing method",
			bindsTo: "http.server.api:/users",
		},
		{
			name:    "invalid method",
			bindsTo: "http.server.api:INVALID:/users",
		},
		{
			name:    "missing path",
			bindsTo: "http.server.api:GET:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			spec := &parser.Spec{
				Components: []parser.Component{
					{
						ID:   "http.server.api",
						Kind: "http.server",
						Spec: map[string]interface{}{
							"framework": "hono",
							"port":      3000,
						},
					},
					{
						ID:   "usecase.test",
						Kind: "usecase",
						Spec: map[string]interface{}{
							"binds_to": tt.bindsTo,
							"goal":     "Test",
						},
					},
				},
			}

			// when
			b := NewBuilder()
			_, errs := b.Build(spec)

			// then
			if len(errs) == 0 {
				t.Error("Build() expected error for invalid binds_to format")
			}
		})
	}
}

func TestBuilder_Build_UsecaseBindsToNonServer(t *testing.T) {
	// given: a spec where usecase binds to non-http.server component
	spec := &parser.Spec{
		Components: []parser.Component{
			{
				ID:   "postgres.primary",
				Kind: "postgres",
				Spec: map[string]interface{}{
					"provider": "drizzle",
				},
			},
			{
				ID:   "usecase.test",
				Kind: "usecase",
				Spec: map[string]interface{}{
					"binds_to": "postgres.primary:GET:/users",
					"goal":     "Test",
				},
			},
		},
	}

	// when
	b := NewBuilder()
	_, errs := b.Build(spec)

	// then: should get error that target is not http.server
	hasTypeError := false
	for _, err := range errs {
		if err != nil && strings.Contains(err.Error(), "not an http.server") {
			hasTypeError = true
			break
		}
	}
	if !hasTypeError {
		t.Error("Build() expected error about binding to non-http.server")
	}
}
