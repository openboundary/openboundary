package typescript

import (
	"strings"
	"testing"

	"github.com/stack-bound/stack-bound/internal/ir"
)

func TestDockerGenerator_Name(t *testing.T) {
	g := NewDockerGenerator()
	if got := g.Name(); got != "typescript-docker" {
		t.Errorf("Name() = %v, want %v", got, "typescript-docker")
	}
}

func TestDockerGenerator_Generate(t *testing.T) {
	tests := []struct {
		name    string
		ir      *ir.IR
		wantErr bool
		checks  func(t *testing.T, files map[string][]byte)
	}{
		{
			name: "generates docker files without postgres",
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
				// Check Dockerfile exists
				if _, ok := files["Dockerfile"]; !ok {
					t.Error("Dockerfile not generated")
				}

				// Check docker-compose.yml exists
				composeContent, ok := files["docker-compose.yml"]
				if !ok {
					t.Error("docker-compose.yml not generated")
				}

				// Should not have postgres service
				if strings.Contains(string(composeContent), "postgres:") {
					t.Error("docker-compose.yml should not contain postgres service")
				}

				// Should have app service
				if !strings.Contains(string(composeContent), "app:") {
					t.Error("docker-compose.yml should contain app service")
				}

				// Check .dockerignore exists
				if _, ok := files[".dockerignore"]; !ok {
					t.Error(".dockerignore not generated")
				}
			},
		},
		{
			name: "generates docker files with postgres",
			ir: &ir.IR{
				Components: map[string]*ir.Component{
					"db": {
						ID:   "db",
						Kind: ir.KindPostgres,
						Postgres: &ir.PostgresSpec{
							Provider: "drizzle",
						},
					},
					"api": {
						ID:   "api",
						Kind: ir.KindHTTPServer,
						HTTPServer: &ir.HTTPServerSpec{
							Port:      8080,
							DependsOn: []string{"db"},
						},
					},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, files map[string][]byte) {
				composeContent := string(files["docker-compose.yml"])

				// Should have postgres service
				if !strings.Contains(composeContent, "postgres:") {
					t.Error("docker-compose.yml should contain postgres service")
				}

				// Should have postgres image
				if !strings.Contains(composeContent, "image: postgres:16-alpine") {
					t.Error("docker-compose.yml should use postgres:16-alpine image")
				}

				// Should have healthcheck
				if !strings.Contains(composeContent, "healthcheck:") {
					t.Error("docker-compose.yml should have healthcheck")
				}

				// Should have DATABASE_URL
				if !strings.Contains(composeContent, "DATABASE_URL:") {
					t.Error("docker-compose.yml should have DATABASE_URL")
				}

				// Should have depends_on with postgres
				if !strings.Contains(composeContent, "depends_on:") {
					t.Error("docker-compose.yml should have depends_on")
				}

				// Should have postgres_data volume
				if !strings.Contains(composeContent, "postgres_data:") {
					t.Error("docker-compose.yml should have postgres_data volume")
				}

				// Should use custom port 8080 (with env var template)
				if !strings.Contains(composeContent, ":-8080}:8080") {
					t.Error("docker-compose.yml should use port 8080")
				}
			},
		},
		{
			name: "dockerfile has multi-stage build",
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
				dockerfile := string(files["Dockerfile"])

				// Should have builder stage
				if !strings.Contains(dockerfile, "FROM node:20-alpine AS builder") {
					t.Error("Dockerfile should have builder stage")
				}

				// Should have production stage
				if !strings.Contains(dockerfile, "FROM node:20-alpine AS production") {
					t.Error("Dockerfile should have production stage")
				}

				// Should copy from builder
				if !strings.Contains(dockerfile, "COPY --from=builder") {
					t.Error("Dockerfile should copy from builder stage")
				}

				// Should have healthcheck
				if !strings.Contains(dockerfile, "HEALTHCHECK") {
					t.Error("Dockerfile should have healthcheck")
				}

				// Should create non-root user
				if !strings.Contains(dockerfile, "USER nodejs") {
					t.Error("Dockerfile should use non-root user")
				}

				// Should expose port
				if !strings.Contains(dockerfile, "EXPOSE 3000") {
					t.Error("Dockerfile should expose port 3000")
				}
			},
		},
		{
			name: "dockerignore excludes correct files",
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
				dockerignore := string(files[".dockerignore"])

				mustHave := []string{
					"node_modules/",
					"dist/",
					".env",
					".git/",
					"coverage/",
					"*.log",
				}

				for _, pattern := range mustHave {
					if !strings.Contains(dockerignore, pattern) {
						t.Errorf(".dockerignore should contain %q", pattern)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDockerGenerator()
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

func TestDockerGenerator_generateDockerCompose_MultipleServers(t *testing.T) {
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
		},
	}

	g := NewDockerGenerator()
	output, err := g.Generate(ir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	compose := string(output.Files["docker-compose.yml"])

	// Should use first server's port (admin=4000, alphabetically first) with env var template
	if !strings.Contains(compose, ":-4000}:4000") {
		t.Error("docker-compose.yml should use first (alphabetically) server port 4000")
	}
}
