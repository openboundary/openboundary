// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package validator

import (
	"testing"

	"github.com/openboundary/openboundary/internal/ir"
	"github.com/openboundary/openboundary/internal/parser"
)

func TestIRValidator_ValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ValidationError
		contains string
	}{
		{
			name:     "with ID",
			err:      ValidationError{ID: "comp.test", Message: "some error"},
			contains: "comp.test: some error",
		},
		{
			name:     "without ID",
			err:      ValidationError{Message: "some error"},
			contains: "some error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.contains {
				t.Errorf("Error() = %q, expected %q", got, tt.contains)
			}
		})
	}
}

func TestIRValidator_ValidSpec(t *testing.T) {
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
				ID:   "postgres.primary",
				Kind: "postgres",
				Spec: map[string]interface{}{
					"provider": "drizzle",
					"schema":   "./schema.ts",
				},
			},
			{
				ID:   "usecase.test",
				Kind: "usecase",
				Spec: map[string]interface{}{
					"binds_to": "http.server.api:POST:/test",
					"goal":     "Test usecase",
				},
			},
		},
	}

	b := ir.NewBuilder()
	builtIR, buildErrs := b.Build(spec)
	if len(buildErrs) > 0 {
		t.Fatalf("Build() errors: %v", buildErrs)
	}

	v := NewIRValidator()
	errs := v.Validate(builtIR)
	if len(errs) != 0 {
		t.Errorf("Validate() returned %d errors, expected 0", len(errs))
		for _, e := range errs {
			t.Logf("  error: %v", e)
		}
	}
}

func TestIRValidator_HTTPServer(t *testing.T) {
	tests := []struct {
		name       string
		spec       map[string]interface{}
		wantErrors int
	}{
		{
			name: "valid",
			spec: map[string]interface{}{
				"framework": "hono",
				"port":      3000,
			},
			wantErrors: 0,
		},
		{
			name: "missing framework",
			spec: map[string]interface{}{
				"port": 3000,
			},
			wantErrors: 1,
		},
		{
			name: "missing port",
			spec: map[string]interface{}{
				"framework": "hono",
			},
			wantErrors: 1,
		},
		{
			name:       "missing both",
			spec:       map[string]interface{}{},
			wantErrors: 2,
		},
		{
			name: "invalid port negative",
			spec: map[string]interface{}{
				"framework": "hono",
				"port":      -1,
			},
			wantErrors: 1,
		},
		{
			name: "invalid port too high",
			spec: map[string]interface{}{
				"framework": "hono",
				"port":      70000,
			},
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &parser.Spec{
				Components: []parser.Component{
					{ID: "http.server.api", Kind: "http.server", Spec: tt.spec},
				},
			}

			b := ir.NewBuilder()
			builtIR, _ := b.Build(spec)
			v := NewIRValidator()
			errs := v.Validate(builtIR)

			if len(errs) != tt.wantErrors {
				t.Errorf("Validate() returned %d errors, expected %d", len(errs), tt.wantErrors)
				for _, e := range errs {
					t.Logf("  error: %v", e)
				}
			}
		})
	}
}

func TestIRValidator_Middleware(t *testing.T) {
	tests := []struct {
		name       string
		spec       map[string]interface{}
		wantErrors int
	}{
		{
			name: "valid better-auth",
			spec: map[string]interface{}{
				"provider": "better-auth",
				"config":   "./auth.ts",
			},
			wantErrors: 0,
		},
		{
			name: "valid casbin",
			spec: map[string]interface{}{
				"provider": "casbin",
				"model":    "./model.conf",
				"policy":   "./policy.csv",
			},
			wantErrors: 0,
		},
		{
			name:       "missing provider",
			spec:       map[string]interface{}{},
			wantErrors: 1,
		},
		{
			name: "better-auth missing config",
			spec: map[string]interface{}{
				"provider": "better-auth",
			},
			wantErrors: 1,
		},
		{
			name: "casbin missing model",
			spec: map[string]interface{}{
				"provider": "casbin",
				"policy":   "./policy.csv",
			},
			wantErrors: 1,
		},
		{
			name: "casbin missing policy",
			spec: map[string]interface{}{
				"provider": "casbin",
				"model":    "./model.conf",
			},
			wantErrors: 1,
		},
		{
			name: "casbin missing both",
			spec: map[string]interface{}{
				"provider": "casbin",
			},
			wantErrors: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &parser.Spec{
				Components: []parser.Component{
					{ID: "middleware.test", Kind: "middleware", Spec: tt.spec},
				},
			}

			b := ir.NewBuilder()
			builtIR, _ := b.Build(spec)
			v := NewIRValidator()
			errs := v.Validate(builtIR)

			if len(errs) != tt.wantErrors {
				t.Errorf("Validate() returned %d errors, expected %d", len(errs), tt.wantErrors)
				for _, e := range errs {
					t.Logf("  error: %v", e)
				}
			}
		})
	}
}

func TestIRValidator_Postgres(t *testing.T) {
	tests := []struct {
		name       string
		spec       map[string]interface{}
		wantErrors int
	}{
		{
			name: "valid",
			spec: map[string]interface{}{
				"provider": "drizzle",
				"schema":   "./schema.ts",
			},
			wantErrors: 0,
		},
		{
			name: "missing provider",
			spec: map[string]interface{}{
				"schema": "./schema.ts",
			},
			wantErrors: 1,
		},
		{
			name: "missing schema",
			spec: map[string]interface{}{
				"provider": "drizzle",
			},
			wantErrors: 1,
		},
		{
			name:       "missing both",
			spec:       map[string]interface{}{},
			wantErrors: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &parser.Spec{
				Components: []parser.Component{
					{ID: "postgres.primary", Kind: "postgres", Spec: tt.spec},
				},
			}

			b := ir.NewBuilder()
			builtIR, _ := b.Build(spec)
			v := NewIRValidator()
			errs := v.Validate(builtIR)

			if len(errs) != tt.wantErrors {
				t.Errorf("Validate() returned %d errors, expected %d", len(errs), tt.wantErrors)
				for _, e := range errs {
					t.Logf("  error: %v", e)
				}
			}
		})
	}
}

func TestIRValidator_Usecase(t *testing.T) {
	baseComponents := []parser.Component{
		{
			ID:   "http.server.api",
			Kind: "http.server",
			Spec: map[string]interface{}{
				"framework": "hono",
				"port":      3000,
			},
		},
	}

	tests := []struct {
		name       string
		spec       map[string]interface{}
		wantErrors int
	}{
		{
			name: "valid",
			spec: map[string]interface{}{
				"binds_to": "http.server.api:POST:/users",
				"goal":     "Create user",
			},
			wantErrors: 0,
		},
		{
			name: "missing binds_to",
			spec: map[string]interface{}{
				"goal": "Create user",
			},
			wantErrors: 1,
		},
		{
			name: "missing goal",
			spec: map[string]interface{}{
				"binds_to": "http.server.api:POST:/users",
			},
			wantErrors: 1,
		},
		{
			name:       "missing both",
			spec:       map[string]interface{}{},
			wantErrors: 2,
		},
		{
			name: "invalid binds_to format",
			spec: map[string]interface{}{
				"binds_to": "invalid",
				"goal":     "Test",
			},
			wantErrors: 1,
		},
		{
			name: "invalid HTTP method",
			spec: map[string]interface{}{
				"binds_to": "http.server.api:INVALID:/users",
				"goal":     "Test",
			},
			wantErrors: 1,
		},
		{
			name: "invalid path no slash",
			spec: map[string]interface{}{
				"binds_to": "http.server.api:GET:users",
				"goal":     "Test",
			},
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := append(baseComponents, parser.Component{
				ID:   "usecase.test",
				Kind: "usecase",
				Spec: tt.spec,
			})

			spec := &parser.Spec{Components: components}

			b := ir.NewBuilder()
			builtIR, _ := b.Build(spec)
			v := NewIRValidator()
			errs := v.Validate(builtIR)

			if len(errs) != tt.wantErrors {
				t.Errorf("Validate() returned %d errors, expected %d", len(errs), tt.wantErrors)
				for _, e := range errs {
					t.Logf("  error: %v", e)
				}
			}
		})
	}
}

func TestIRValidator_MiddlewareTypeCheck(t *testing.T) {
	spec := &parser.Spec{
		Components: []parser.Component{
			{
				ID:   "http.server.api",
				Kind: "http.server",
				Spec: map[string]interface{}{
					"framework":  "hono",
					"port":       3000,
					"middleware": []interface{}{"postgres.primary"},
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
		},
	}

	b := ir.NewBuilder()
	builtIR, _ := b.Build(spec)
	v := NewIRValidator()
	errs := v.Validate(builtIR)

	if len(errs) != 1 {
		t.Errorf("Validate() returned %d errors, expected 1 (wrong middleware type)", len(errs))
		for _, e := range errs {
			t.Logf("  error: %v", e)
		}
	}
}

func TestIRValidator_BindsToTypeCheck(t *testing.T) {
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
			{
				ID:   "usecase.test",
				Kind: "usecase",
				Spec: map[string]interface{}{
					"binds_to": "postgres.primary:GET:/test",
					"goal":     "Test",
				},
			},
		},
	}

	b := ir.NewBuilder()
	builtIR, _ := b.Build(spec)
	v := NewIRValidator()
	errs := v.Validate(builtIR)

	found := false
	for _, e := range errs {
		if e.ID == "usecase.test" {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should error on binds_to pointing to postgres")
	}
}

func TestIRValidator_NilSpecs(t *testing.T) {
	builtIR := ir.New(&parser.Spec{})

	builtIR.Components["http.server.api"] = &ir.Component{ID: "http.server.api", Kind: ir.KindHTTPServer, HTTPServer: nil}
	builtIR.Components["middleware.auth"] = &ir.Component{ID: "middleware.auth", Kind: ir.KindMiddleware, Middleware: nil}
	builtIR.Components["postgres.db"] = &ir.Component{ID: "postgres.db", Kind: ir.KindPostgres, Postgres: nil}
	builtIR.Components["usecase.test"] = &ir.Component{ID: "usecase.test", Kind: ir.KindUsecase, Usecase: nil}

	v := NewIRValidator()
	errs := v.Validate(builtIR)
	if len(errs) != 4 {
		t.Errorf("Validate() returned %d errors, expected 4 (one for each nil spec)", len(errs))
	}
}

func TestIRValidator_WithCycle(t *testing.T) {
	builtIR := ir.New(&parser.Spec{})

	compA := &ir.Component{
		ID:   "middleware.a",
		Kind: ir.KindMiddleware,
		Middleware: &ir.MiddlewareSpec{
			Provider: "better-auth",
			Config:   "./auth.ts",
		},
		Dependencies: []*ir.Component{},
	}
	compB := &ir.Component{
		ID:   "middleware.b",
		Kind: ir.KindMiddleware,
		Middleware: &ir.MiddlewareSpec{
			Provider: "better-auth",
			Config:   "./auth.ts",
		},
		Dependencies: []*ir.Component{},
	}

	compA.Dependencies = append(compA.Dependencies, compB)
	compB.Dependencies = append(compB.Dependencies, compA)

	builtIR.Components["middleware.a"] = compA
	builtIR.Components["middleware.b"] = compB
	_ = builtIR.Symbols.Define("middleware.a", ir.KindMiddleware, compA)
	_ = builtIR.Symbols.Define("middleware.b", ir.KindMiddleware, compB)

	v := NewIRValidator()
	errs := v.Validate(builtIR)
	if len(errs) == 0 {
		t.Error("Validate() should detect cycle")
	}
}

func TestIRValidator_UsecaseMiddlewareTypeCheck(t *testing.T) {
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
				ID:   "postgres.primary",
				Kind: "postgres",
				Spec: map[string]interface{}{
					"provider": "drizzle",
					"schema":   "./schema.ts",
				},
			},
			{
				ID:   "usecase.test",
				Kind: "usecase",
				Spec: map[string]interface{}{
					"binds_to":   "http.server.api:POST:/test",
					"goal":       "Test",
					"middleware": []interface{}{"postgres.primary"},
				},
			},
		},
	}

	b := ir.NewBuilder()
	builtIR, _ := b.Build(spec)
	v := NewIRValidator()
	errs := v.Validate(builtIR)

	found := false
	for _, e := range errs {
		if e.ID == "usecase.test" {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should error on usecase middleware pointing to postgres")
	}
}

func TestIRValidator_AllHTTPMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
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
							"binds_to": "http.server.api:" + method + ":/test",
							"goal":     "Test",
						},
					},
				},
			}

			b := ir.NewBuilder()
			builtIR, _ := b.Build(spec)
			v := NewIRValidator()
			errs := v.Validate(builtIR)

			for _, e := range errs {
				if e.ID == "usecase.test" {
					t.Errorf("Validate() should accept %s method, got error: %v", method, e)
				}
			}
		})
	}
}
