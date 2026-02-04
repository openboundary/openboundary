package typescript

import (
	"strings"
	"testing"

	"github.com/stack-bound/stack-bound/internal/ir"
)

func TestE2ETestGenerator_Name(t *testing.T) {
	g := NewE2ETestGenerator()
	if got := g.Name(); got != "typescript-e2e" {
		t.Errorf("Name() = %v, want %v", got, "typescript-e2e")
	}
}

func TestE2ETestGenerator_Generate(t *testing.T) {
	tests := []struct {
		name    string
		ir      *ir.IR
		wantErr bool
		checks  func(t *testing.T, files map[string][]byte)
	}{
		{
			name: "generates e2e tests for single server",
			ir: &ir.IR{
				Components: map[string]*ir.Component{
					"api": {
						ID:   "api",
						Kind: ir.KindHTTPServer,
						HTTPServer: &ir.HTTPServerSpec{
							Port: 3000,
						},
					},
					"uc1": {
						ID:   "uc1",
						Kind: ir.KindUsecase,
						Usecase: &ir.UsecaseSpec{
							Binding: &ir.Binding{
								ServerID: "api",
								Method:   "GET",
								Path:     "/users",
							},
						},
					},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, files map[string][]byte) {
				// Check E2E test file exists
				testFile, ok := files["e2e/api.spec.ts"]
				if !ok {
					t.Error("e2e/api.spec.ts not generated")
				}

				testContent := string(testFile)

				// Should have health check test
				if !strings.Contains(testContent, "GET /health") {
					t.Error("E2E test should have health check test")
				}

				// Should have usecase test
				if !strings.Contains(testContent, "GET /users") {
					t.Error("E2E test should have GET /users test")
				}

				// Should use correct base URL
				if !strings.Contains(testContent, "http://localhost:3000") {
					t.Error("E2E test should use correct base URL")
				}

				// Check Playwright config exists
				if _, ok := files["playwright.config.ts"]; !ok {
					t.Error("playwright.config.ts not generated")
				}

				// Check helpers exist
				if _, ok := files["e2e/helpers/setup.ts"]; !ok {
					t.Error("e2e/helpers/setup.ts not generated")
				}
			},
		},
		{
			name: "generates e2e tests with path parameters",
			ir: &ir.IR{
				Components: map[string]*ir.Component{
					"api": {
						ID:   "api",
						Kind: ir.KindHTTPServer,
						HTTPServer: &ir.HTTPServerSpec{
							Port: 3000,
						},
					},
					"uc1": {
						ID:   "uc1",
						Kind: ir.KindUsecase,
						Usecase: &ir.UsecaseSpec{
							Binding: &ir.Binding{
								ServerID: "api",
								Method:   "GET",
								Path:     "/users/{id}",
							},
						},
					},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, files map[string][]byte) {
				testContent := string(files["e2e/api.spec.ts"])

				// Should convert {id} to test value
				if !strings.Contains(testContent, "/users/test-id") {
					t.Error("E2E test should convert path params to test values")
				}
			},
		},
		{
			name: "generates e2e tests with authentication",
			ir: &ir.IR{
				Components: map[string]*ir.Component{
					"auth-mw": {
						ID:   "auth-mw",
						Kind: ir.KindMiddleware,
						Middleware: &ir.MiddlewareSpec{
							Provider: "better-auth",
						},
					},
					"api": {
						ID:   "api",
						Kind: ir.KindHTTPServer,
						HTTPServer: &ir.HTTPServerSpec{
							Port:       3000,
							Middleware: []string{"auth-mw"},
						},
					},
					"uc1": {
						ID:   "uc1",
						Kind: ir.KindUsecase,
						Usecase: &ir.UsecaseSpec{
							Binding: &ir.Binding{
								ServerID: "api",
								Method:   "GET",
								Path:     "/users",
							},
						},
					},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, files map[string][]byte) {
				testContent := string(files["e2e/api.spec.ts"])

				// Should import auth helper
				if !strings.Contains(testContent, "import { createAuthToken }") {
					t.Error("E2E test should import createAuthToken for auth")
				}

				// Should use auth token
				if !strings.Contains(testContent, "createAuthToken") {
					t.Error("E2E test should use createAuthToken")
				}

				if !strings.Contains(testContent, "Authorization") {
					t.Error("E2E test should set Authorization header")
				}

				// Check setup helpers
				setupContent := string(files["e2e/helpers/setup.ts"])
				if !strings.Contains(setupContent, "export function createAuthToken") {
					t.Error("setup.ts should export createAuthToken")
				}
			},
		},
		{
			name: "generates e2e tests for POST/PUT methods",
			ir: &ir.IR{
				Components: map[string]*ir.Component{
					"api": {
						ID:   "api",
						Kind: ir.KindHTTPServer,
						HTTPServer: &ir.HTTPServerSpec{
							Port: 3000,
						},
					},
					"create": {
						ID:   "create",
						Kind: ir.KindUsecase,
						Usecase: &ir.UsecaseSpec{
							Binding: &ir.Binding{
								ServerID: "api",
								Method:   "POST",
								Path:     "/users",
							},
						},
					},
					"update": {
						ID:   "update",
						Kind: ir.KindUsecase,
						Usecase: &ir.UsecaseSpec{
							Binding: &ir.Binding{
								ServerID: "api",
								Method:   "PUT",
								Path:     "/users/{id}",
							},
						},
					},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, files map[string][]byte) {
				testContent := string(files["e2e/api.spec.ts"])

				// Should have POST test with data
				if !strings.Contains(testContent, "POST /users") {
					t.Error("E2E test should have POST /users test")
				}

				// Should have PUT test with data
				if !strings.Contains(testContent, "PUT /users/{id}") {
					t.Error("E2E test should have PUT /users/{id} test")
				}

				// Should include data in request
				if !strings.Contains(testContent, "data: {}") {
					t.Error("E2E test should include data in POST/PUT requests")
				}
			},
		},
		{
			name: "playwright config uses correct port",
			ir: &ir.IR{
				Components: map[string]*ir.Component{
					"api": {
						ID:   "api",
						Kind: ir.KindHTTPServer,
						HTTPServer: &ir.HTTPServerSpec{
							Port: 8080,
						},
					},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, files map[string][]byte) {
				configContent := string(files["playwright.config.ts"])

				// Should use custom port
				if !strings.Contains(configContent, "http://localhost:8080") {
					t.Error("Playwright config should use custom port 8080")
				}

				// Should have webServer config
				if !strings.Contains(configContent, "webServer:") {
					t.Error("Playwright config should have webServer config")
				}

				// Should have retry config for CI
				if !strings.Contains(configContent, "retries:") {
					t.Error("Playwright config should have retries config")
				}
			},
		},
		{
			name: "generates helpers with test utilities",
			ir: &ir.IR{
				Components: map[string]*ir.Component{
					"api": {
						ID:   "api",
						Kind: ir.KindHTTPServer,
						HTTPServer: &ir.HTTPServerSpec{
							Port: 3000,
						},
					},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, files map[string][]byte) {
				setupContent := string(files["e2e/helpers/setup.ts"])

				// Should have test data factory
				if !strings.Contains(setupContent, "export function createTestData") {
					t.Error("setup.ts should export createTestData")
				}

				// Should have wait utility
				if !strings.Contains(setupContent, "export async function waitForCondition") {
					t.Error("setup.ts should export waitForCondition")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewE2ETestGenerator()
			output, err := g.Generate(tt.ir)

			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checks != nil {
				tt.checks(t, output.Files)
			}
		})
	}
}

func TestE2ETestGenerator_MultipleServers(t *testing.T) {
	ir := &ir.IR{
		Components: map[string]*ir.Component{
			"api": {
				ID:   "api",
				Kind: ir.KindHTTPServer,
				HTTPServer: &ir.HTTPServerSpec{
					Port: 3000,
				},
			},
			"admin": {
				ID:   "admin",
				Kind: ir.KindHTTPServer,
				HTTPServer: &ir.HTTPServerSpec{
					Port: 4000,
				},
			},
			"uc1": {
				ID:   "uc1",
				Kind: ir.KindUsecase,
				Usecase: &ir.UsecaseSpec{
					Binding: &ir.Binding{
						ServerID: "api",
						Method:   "GET",
						Path:     "/users",
					},
				},
			},
			"uc2": {
				ID:   "uc2",
				Kind: ir.KindUsecase,
				Usecase: &ir.UsecaseSpec{
					Binding: &ir.Binding{
						ServerID: "admin",
						Method:   "GET",
						Path:     "/settings",
					},
				},
			},
		},
	}

	g := NewE2ETestGenerator()
	output, err := g.Generate(ir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Should generate separate test files
	if _, ok := output.Files["e2e/api.spec.ts"]; !ok {
		t.Error("Should generate e2e/api.spec.ts")
	}

	if _, ok := output.Files["e2e/admin.spec.ts"]; !ok {
		t.Error("Should generate e2e/admin.spec.ts")
	}

	// Check test contents are scoped correctly
	apiTest := string(output.Files["e2e/api.spec.ts"])
	if !strings.Contains(apiTest, "GET /users") {
		t.Error("api.spec.ts should contain GET /users")
	}
	if strings.Contains(apiTest, "GET /settings") {
		t.Error("api.spec.ts should not contain admin routes")
	}

	adminTest := string(output.Files["e2e/admin.spec.ts"])
	if !strings.Contains(adminTest, "GET /settings") {
		t.Error("admin.spec.ts should contain GET /settings")
	}
	if strings.Contains(adminTest, "GET /users") {
		t.Error("admin.spec.ts should not contain api routes")
	}
}
