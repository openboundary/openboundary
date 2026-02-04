package ir

import (
	"testing"

	"github.com/openboundary/openboundary/internal/parser"
)

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		contains string
	}{
		{
			name:     "with ID",
			err:      &ValidationError{ID: "comp.test", Message: "some error"},
			contains: "comp.test: some error",
		},
		{
			name:     "without ID",
			err:      &ValidationError{Message: "some error"},
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

func TestIR_Validate_ValidSpec(t *testing.T) {
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

	b := NewBuilder()
	ir, buildErrs := b.Build(spec)
	if len(buildErrs) > 0 {
		t.Fatalf("Build() errors: %v", buildErrs)
	}

	errs := ir.Validate()
	if len(errs) != 0 {
		t.Errorf("Validate() returned %d errors, expected 0", len(errs))
		for _, e := range errs {
			t.Logf("  error: %v", e)
		}
	}
}

func TestIR_Validate_HTTPServer(t *testing.T) {
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

			b := NewBuilder()
			ir, _ := b.Build(spec)
			errs := ir.Validate()

			if len(errs) != tt.wantErrors {
				t.Errorf("Validate() returned %d errors, expected %d", len(errs), tt.wantErrors)
				for _, e := range errs {
					t.Logf("  error: %v", e)
				}
			}
		})
	}
}

func TestIR_Validate_Middleware(t *testing.T) {
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

			b := NewBuilder()
			ir, _ := b.Build(spec)
			errs := ir.Validate()

			if len(errs) != tt.wantErrors {
				t.Errorf("Validate() returned %d errors, expected %d", len(errs), tt.wantErrors)
				for _, e := range errs {
					t.Logf("  error: %v", e)
				}
			}
		})
	}
}

func TestIR_Validate_Postgres(t *testing.T) {
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

			b := NewBuilder()
			ir, _ := b.Build(spec)
			errs := ir.Validate()

			if len(errs) != tt.wantErrors {
				t.Errorf("Validate() returned %d errors, expected %d", len(errs), tt.wantErrors)
				for _, e := range errs {
					t.Logf("  error: %v", e)
				}
			}
		})
	}
}

func TestIR_Validate_Usecase(t *testing.T) {
	// First create an http.server for binds_to references
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

			b := NewBuilder()
			ir, _ := b.Build(spec)
			errs := ir.Validate()

			if len(errs) != tt.wantErrors {
				t.Errorf("Validate() returned %d errors, expected %d", len(errs), tt.wantErrors)
				for _, e := range errs {
					t.Logf("  error: %v", e)
				}
			}
		})
	}
}

func TestIR_Validate_MiddlewareTypeCheck(t *testing.T) {
	spec := &parser.Spec{
		Components: []parser.Component{
			{
				ID:   "http.server.api",
				Kind: "http.server",
				Spec: map[string]interface{}{
					"framework":  "hono",
					"port":       3000,
					"middleware": []interface{}{"postgres.primary"}, // Wrong type
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

	b := NewBuilder()
	ir, _ := b.Build(spec)
	errs := ir.Validate()

	if len(errs) != 1 {
		t.Errorf("Validate() returned %d errors, expected 1 (wrong middleware type)", len(errs))
		for _, e := range errs {
			t.Logf("  error: %v", e)
		}
	}
}

func TestIR_Validate_BindsToTypeCheck(t *testing.T) {
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
					"binds_to": "postgres.primary:GET:/test", // Wrong type
					"goal":     "Test",
				},
			},
		},
	}

	b := NewBuilder()
	ir, _ := b.Build(spec)
	errs := ir.Validate()

	found := false
	for _, e := range errs {
		if ve, ok := e.(*ValidationError); ok && ve.ID == "usecase.test" {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should error on binds_to pointing to postgres")
	}
}

func TestValidateBindsTo(t *testing.T) {
	tests := []struct {
		bindsTo string
		wantErr bool
	}{
		{"server:GET:/path", false},
		{"server:POST:/path", false},
		{"server:PUT:/path", false},
		{"server:PATCH:/path", false},
		{"server:DELETE:/path", false},
		{"server:INVALID:/path", true},
		{"server:GET:path", true}, // no leading /
		{"serveronly", true},
		{"server:GET", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.bindsTo, func(t *testing.T) {
			err := validateBindsTo(tt.bindsTo)
			if tt.wantErr && err == nil {
				t.Error("validateBindsTo() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateBindsTo() unexpected error: %v", err)
			}
		})
	}
}

func TestIR_Validate_NilSpecs(t *testing.T) {
	// Test that validation handles nil specs gracefully
	ir := New(&parser.Spec{})

	// Manually create components with nil specs
	ir.Components["http.server.api"] = &Component{ID: "http.server.api", Kind: KindHTTPServer, HTTPServer: nil}
	ir.Components["middleware.auth"] = &Component{ID: "middleware.auth", Kind: KindMiddleware, Middleware: nil}
	ir.Components["postgres.db"] = &Component{ID: "postgres.db", Kind: KindPostgres, Postgres: nil}
	ir.Components["usecase.test"] = &Component{ID: "usecase.test", Kind: KindUsecase, Usecase: nil}

	errs := ir.Validate()
	if len(errs) != 4 {
		t.Errorf("Validate() returned %d errors, expected 4 (one for each nil spec)", len(errs))
	}
}

func TestIR_Validate_WithCycle(t *testing.T) {
	// Test that validation catches cycles
	ir := New(&parser.Spec{})

	compA := &Component{
		ID:   "middleware.a",
		Kind: KindMiddleware,
		Middleware: &MiddlewareSpec{
			Provider: "better-auth",
			Config:   "./auth.ts",
		},
		Dependencies: []*Component{},
	}
	compB := &Component{
		ID:   "middleware.b",
		Kind: KindMiddleware,
		Middleware: &MiddlewareSpec{
			Provider: "better-auth",
			Config:   "./auth.ts",
		},
		Dependencies: []*Component{},
	}

	// Create cycle
	compA.Dependencies = append(compA.Dependencies, compB)
	compB.Dependencies = append(compB.Dependencies, compA)

	ir.Components["middleware.a"] = compA
	ir.Components["middleware.b"] = compB
	_ = ir.Symbols.Define("middleware.a", KindMiddleware, compA)
	_ = ir.Symbols.Define("middleware.b", KindMiddleware, compB)

	errs := ir.Validate()
	found := false
	for _, e := range errs {
		if ve, ok := e.(*ValidationError); ok {
			if ve.Message != "" && len(ve.Message) > 0 {
				found = true
			}
		}
	}
	if !found && len(errs) == 0 {
		t.Error("Validate() should detect cycle")
	}
}

func TestIR_Validate_UsecaseMiddlewareTypeCheck(t *testing.T) {
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
					"middleware": []interface{}{"postgres.primary"}, // Wrong type!
				},
			},
		},
	}

	b := NewBuilder()
	ir, _ := b.Build(spec)
	errs := ir.Validate()

	found := false
	for _, e := range errs {
		if ve, ok := e.(*ValidationError); ok && ve.ID == "usecase.test" {
			found = true
		}
	}
	if !found {
		t.Error("Validate() should error on usecase middleware pointing to postgres")
	}
}

func TestIR_Validate_AllHTTPMethods(t *testing.T) {
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

			b := NewBuilder()
			ir, _ := b.Build(spec)
			errs := ir.Validate()

			for _, e := range errs {
				if ve, ok := e.(*ValidationError); ok && ve.ID == "usecase.test" {
					t.Errorf("Validate() should accept %s method, got error: %v", method, ve)
				}
			}
		})
	}
}
