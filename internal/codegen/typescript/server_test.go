package typescript

import (
	"strings"
	"testing"

	"github.com/openboundary/openboundary/internal/codegen"
	"github.com/openboundary/openboundary/internal/ir"
	"github.com/openboundary/openboundary/internal/parser"
)

func TestNewHonoServerGenerator(t *testing.T) {
	// given/when
	g := NewHonoServerGenerator()

	// then
	if g == nil {
		t.Fatal("NewHonoServerGenerator() returned nil")
	}
}

func TestHonoServerGenerator_Name(t *testing.T) {
	// given
	g := NewHonoServerGenerator()

	// when
	name := g.Name()

	// then
	if name != "typescript-hono" {
		t.Errorf("Name() = %q, want %q", name, "typescript-hono")
	}
}

func TestHonoServerGenerator_Generate_Index(t *testing.T) {
	// given: IR with http.server
	i := createTestIR()

	// when
	g := NewHonoServerGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	indexContent, ok := output.Files["src/index.ts"]
	if !ok {
		t.Fatal("src/index.ts not found in output")
	}

	content := string(indexContent)
	if !strings.Contains(content, "serve") {
		t.Error("index.ts should import serve from @hono/node-server")
	}
	if !strings.Contains(content, "main()") {
		t.Error("index.ts should have main function")
	}
}

func TestHonoServerGenerator_Generate_ServerFile(t *testing.T) {
	// given: IR with http.server
	i := createTestIR()

	// when
	g := NewHonoServerGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	serverContent, ok := output.Files["src/components/servers/http-server-api.ts"]
	if !ok {
		t.Fatal("server file not found in output")
	}

	content := string(serverContent)
	if !strings.Contains(content, "createHttpServerApiApp") {
		t.Error("server file should have createHttpServerApiApp function")
	}
	if !strings.Contains(content, "Hono") {
		t.Error("server file should import Hono")
	}
}

func TestHonoServerGenerator_Generate_Routes(t *testing.T) {
	// given: IR with http.server and usecases
	i := createTestIR()

	// when
	g := NewHonoServerGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(output.Files["src/components/servers/http-server-api.ts"])

	// Check for POST route
	if !strings.Contains(content, "app.post('/users'") {
		t.Error("server should have POST /users route")
	}

	// Check for GET route with param
	if !strings.Contains(content, "app.get('/users/:id'") {
		t.Error("server should have GET /users/:id route")
	}
}

func TestHonoServerGenerator_Generate_MiddlewareFile(t *testing.T) {
	// given: IR with middleware
	i := createTestIR()

	// when
	g := NewHonoServerGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	mwContent, ok := output.Files["src/components/middlewares/middleware-authn.ts"]
	if !ok {
		t.Fatal("middleware file not found in output")
	}

	content := string(mwContent)
	if !strings.Contains(content, "createMiddleware") {
		t.Error("middleware file should use createMiddleware")
	}
	if !strings.Contains(content, "middlewareAuthnMiddleware") {
		t.Error("middleware file should export middleware function")
	}
}

func TestHonoServerGenerator_Generate_PostgresClient(t *testing.T) {
	// given: IR with postgres
	i := createTestIR()

	// when
	g := NewHonoServerGenerator()
	output, err := g.Generate(i)

	// then
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	pgContent, ok := output.Files["src/components/postgres/postgres-primary.ts"]
	if !ok {
		t.Fatal("postgres client file not found in output")
	}

	content := string(pgContent)
	if !strings.Contains(content, "drizzle") {
		t.Error("postgres file should import drizzle")
	}
	if !strings.Contains(content, "createPostgresPrimaryClient") {
		t.Error("postgres file should export create client function")
	}
}

func TestHonoServerGenerator_ImplementsGenerator(t *testing.T) {
	// given
	g := NewHonoServerGenerator()

	// then
	var _ codegen.Generator = g
}

func TestConvertPathParams(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/users", "/users"},
		{"/users/{id}", "/users/:id"},
		{"/users/{userId}/posts/{postId}", "/users/:userId/posts/:postId"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// when
			got := convertPathParams(tt.input)

			// then
			if got != tt.want {
				t.Errorf("convertPathParams(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractPathParams(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"/users", nil},
		{"/users/{id}", []string{"id"}},
		{"/users/{userId}/posts/{postId}", []string{"userId", "postId"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// when
			got := extractPathParams(tt.input)

			// then
			if len(got) != len(tt.want) {
				t.Errorf("extractPathParams(%q) = %v, want %v", tt.input, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("extractPathParams(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestToFunctionName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"usecase.create-user", "createUserUsecase"},
		{"usecase.get-user", "getUserUsecase"},
		{"usecase.list-users", "listUsersUsecase"},
		{"usecase.delete-user", "deleteUserUsecase"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// when
			got := toFunctionName(tt.input)

			// then
			if got != tt.want {
				t.Errorf("toFunctionName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"http.server.api", "httpServerApi"},
		{"middleware.authn", "middlewareAuthn"},
		{"postgres.primary", "postgresPrimary"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// when
			got := toCamelCase(tt.input)

			// then
			if got != tt.want {
				t.Errorf("toCamelCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"http.server.api", "HttpServerApi"},
		{"middleware.authn", "MiddlewareAuthn"},
		{"postgres.primary", "PostgresPrimary"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// when
			got := toPascalCase(tt.input)

			// then
			if got != tt.want {
				t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// Helper to create a test IR
func createTestIR() *ir.IR {
	postgres := &ir.Component{
		ID:   "postgres.primary",
		Kind: ir.KindPostgres,
		Postgres: &ir.PostgresSpec{
			Provider: "drizzle",
			Schema:   "./src/db/schema.ts",
		},
	}

	authn := &ir.Component{
		ID:   "middleware.authn",
		Kind: ir.KindMiddleware,
		Middleware: &ir.MiddlewareSpec{
			Provider: "better-auth",
			Config:   "./auth.config.ts",
		},
	}

	authz := &ir.Component{
		ID:   "middleware.authz",
		Kind: ir.KindMiddleware,
		Middleware: &ir.MiddlewareSpec{
			Provider: "casbin",
			Model:    "./model.conf",
			Policy:   "./policy.csv",
		},
		Dependencies: []*ir.Component{authn},
	}

	server := &ir.Component{
		ID:   "http.server.api",
		Kind: ir.KindHTTPServer,
		HTTPServer: &ir.HTTPServerSpec{
			Framework:  "hono",
			Port:       3000,
			Middleware: []string{"middleware.authn", "middleware.authz"},
			DependsOn:  []string{"postgres.primary"},
		},
		Dependencies: []*ir.Component{postgres, authn, authz},
	}

	createUser := &ir.Component{
		ID:   "usecase.create-user",
		Kind: ir.KindUsecase,
		Usecase: &ir.UsecaseSpec{
			BindsTo:    "http.server.api:POST:/users",
			Middleware: []string{},
			Goal:       "Create a new user",
			Binding: &ir.Binding{
				ServerID: "http.server.api",
				Method:   "POST",
				Path:     "/users",
			},
		},
	}

	getUser := &ir.Component{
		ID:   "usecase.get-user",
		Kind: ir.KindUsecase,
		Usecase: &ir.UsecaseSpec{
			BindsTo:    "http.server.api:GET:/users/{id}",
			Middleware: []string{"middleware.authn", "middleware.authz"},
			Goal:       "Get user by ID",
			Binding: &ir.Binding{
				ServerID: "http.server.api",
				Method:   "GET",
				Path:     "/users/{id}",
			},
		},
	}

	return &ir.IR{
		Spec: &parser.Spec{
			Name:    "test-api",
			Version: "1.0.0",
		},
		Components: map[string]*ir.Component{
			"http.server.api":     server,
			"middleware.authn":    authn,
			"middleware.authz":    authz,
			"postgres.primary":    postgres,
			"usecase.create-user": createUser,
			"usecase.get-user":    getUser,
		},
	}
}
