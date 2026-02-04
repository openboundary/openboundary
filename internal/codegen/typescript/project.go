package typescript

import (
	"encoding/json"
	"fmt"

	"github.com/stack-bound/stack-bound/internal/codegen"
	"github.com/stack-bound/stack-bound/internal/ir"
)

// ProjectGenerator generates project configuration files.
type ProjectGenerator struct{}

// NewProjectGenerator creates a new project generator.
func NewProjectGenerator() *ProjectGenerator {
	return &ProjectGenerator{}
}

// Name returns the generator name.
func (g *ProjectGenerator) Name() string {
	return "typescript-project"
}

// PackageJSON represents the package.json structure.
type PackageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Description     string            `json:"description,omitempty"`
	Type            string            `json:"type"`
	Main            string            `json:"main"`
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// TSConfig represents the tsconfig.json structure.
type TSConfig struct {
	CompilerOptions TSConfigCompilerOptions `json:"compilerOptions"`
	Include         []string                `json:"include"`
	Exclude         []string                `json:"exclude"`
}

// TSConfigCompilerOptions represents TypeScript compiler options.
type TSConfigCompilerOptions struct {
	Target            string `json:"target"`
	Module            string `json:"module"`
	ModuleResolution  string `json:"moduleResolution"`
	Strict            bool   `json:"strict"`
	ESModuleInterop   bool   `json:"esModuleInterop"`
	SkipLibCheck      bool   `json:"skipLibCheck"`
	ForceConsistentCasingInFileNames bool `json:"forceConsistentCasingInFileNames"`
	OutDir            string `json:"outDir"`
	RootDir           string `json:"rootDir"`
	Declaration       bool   `json:"declaration"`
	ResolveJsonModule bool   `json:"resolveJsonModule"`
}

// OrvalConfig represents the orval.config.ts configuration.
type OrvalConfig struct {
	OutputPath string
	InputPath  string
}

// Generate produces project configuration files.
func (g *ProjectGenerator) Generate(i *ir.IR) (*codegen.Output, error) {
	output := codegen.NewOutput()

	// Generate package.json
	pkgJSON, err := g.generatePackageJSON(i)
	if err != nil {
		return nil, fmt.Errorf("failed to generate package.json: %w", err)
	}
	output.AddFile("package.json", pkgJSON)

	// Generate tsconfig.json
	tsConfig, err := g.generateTSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to generate tsconfig.json: %w", err)
	}
	output.AddFile("tsconfig.json", tsConfig)

	// Generate orval.config.ts for each server with OpenAPI
	for _, comp := range i.Components {
		if comp.Kind != ir.KindHTTPServer || comp.HTTPServer == nil {
			continue
		}
		if comp.HTTPServer.OpenAPI == "" {
			continue
		}

		orvalConfig := g.generateOrvalConfig(comp)
		output.AddFile("orval.config.ts", []byte(orvalConfig))
		break // Only one orval config needed
	}

	// Generate .gitignore
	output.AddFile(".gitignore", []byte(gitignoreContent))

	return output, nil
}

func (g *ProjectGenerator) generatePackageJSON(i *ir.IR) ([]byte, error) {
	// Determine dependencies based on components
	deps := map[string]string{
		"hono":              "^4.0.0",
		"@hono/node-server": "^1.13.0",
	}
	devDeps := map[string]string{
		"typescript":      "^5.0.0",
		"@types/node":     "^20.0.0",
		"vitest":          "^2.0.0",
		"orval":           "^7.0.0",
		"tsx":             "^4.0.0",
		"@playwright/test": "^1.42.0",
	}

	// Add dependencies based on component types
	for _, comp := range i.Components {
		switch comp.Kind {
		case ir.KindPostgres:
			if comp.Postgres != nil && comp.Postgres.Provider == "drizzle" {
				deps["drizzle-orm"] = "^0.41.0"
				deps["postgres"] = "^3.4.0"
				devDeps["drizzle-kit"] = "^0.31.0"
			}
		case ir.KindMiddleware:
			if comp.Middleware != nil {
				switch comp.Middleware.Provider {
				case "better-auth":
					deps["better-auth"] = "^1.4.0"
				case "casbin":
					deps["casbin"] = "^5.0.0"
				}
			}
		}
	}

	name := "generated-api"
	version := "0.0.1"
	description := ""
	if i.Spec != nil {
		if i.Spec.Name != "" {
			name = i.Spec.Name
		}
		if i.Spec.Version != "" {
			version = i.Spec.Version
		}
		if i.Spec.Description != "" {
			description = i.Spec.Description
		}
	}

	scripts := map[string]string{
		"build":          "tsc",
		"dev":            "tsx watch src/index.ts",
		"start":          "node dist/index.js",
		"test":           "vitest run",
		"test:watch":     "vitest",
		"test:e2e":       "playwright test",
		"test:e2e:ui":    "playwright test --ui",
		"generate:types": "orval",
		"lint":           "tsc --noEmit",
		"docker:build":   "docker build -t app .",
		"docker:up":      "docker-compose up -d",
		"docker:down":    "docker-compose down",
		"docker:logs":    "docker-compose logs -f",
		"docker:ps":      "docker-compose ps",
		"docker:clean":   "docker-compose down -v",
	}

	// Add conditional database scripts if postgres is present
	for _, comp := range i.Components {
		if comp.Kind == ir.KindPostgres && comp.Postgres != nil {
			if comp.Postgres.Provider == "drizzle" {
				scripts["db:migrate"] = "drizzle-kit migrate"
				scripts["db:push"] = "drizzle-kit push"
				scripts["db:studio"] = "drizzle-kit studio"
			}
			break
		}
	}

	pkg := PackageJSON{
		Name:            name,
		Version:         version,
		Description:     description,
		Type:            "module",
		Main:            "dist/index.js",
		Scripts:         scripts,
		Dependencies:    deps,
		DevDependencies: devDeps,
	}

	return json.MarshalIndent(pkg, "", "  ")
}

func (g *ProjectGenerator) generatePackageLock(i *ir.IR) ([]byte, error) {
	// Get package name and version
	name := "generated-api"
	version := "0.1.0"
	if i.Spec != nil {
		if i.Spec.Name != "" {
			name = i.Spec.Name
		}
		if i.Spec.Version != "" {
			version = i.Spec.Version
		}
	}

	// Create minimal package-lock.json v3 format
	// This ensures npm ci works but actual dependency resolution happens at install time
	lockFile := map[string]interface{}{
		"name":            name,
		"version":         version,
		"lockfileVersion": 3,
		"requires":        true,
		"packages": map[string]interface{}{
			"": map[string]interface{}{
				"name":    name,
				"version": version,
			},
		},
	}

	return json.MarshalIndent(lockFile, "", "  ")
}

func (g *ProjectGenerator) generateTSConfig() ([]byte, error) {
	config := TSConfig{
		CompilerOptions: TSConfigCompilerOptions{
			Target:            "ES2022",
			Module:            "ESNext",
			ModuleResolution:  "bundler",
			Strict:            true,
			ESModuleInterop:   true,
			SkipLibCheck:      true,
			ForceConsistentCasingInFileNames: true,
			OutDir:            "./dist",
			RootDir:           "./src",
			Declaration:       true,
			ResolveJsonModule: true,
		},
		Include: []string{"src/**/*"},
		Exclude: []string{"node_modules", "dist"},
	}

	return json.MarshalIndent(config, "", "  ")
}

func (g *ProjectGenerator) generateOrvalConfig(server *ir.Component) string {
	// OpenAPI spec is colocated with the server component
	// Generate schemas colocated with usecases
	serverFilename := sanitizeFilename(server.ID)
	return fmt.Sprintf(`import { defineConfig } from 'orval';

export default defineConfig({
  api: {
    input: './src/components/servers/%s.schema.yaml',
    output: {
      mode: 'single',
      target: './src/components/usecases/schemas.ts',
      client: 'fetch',
      override: {
        // Only generate types, not implementation
        mutator: undefined,
      },
    },
  },
});
`, serverFilename)
}

const gitignoreContent = `# Dependencies
node_modules/

# Build output
dist/

# Environment
.env
.env.local
.env.*.local

# IDE
.vscode/
.idea/

# OS
.DS_Store
Thumbs.db

# Test coverage
coverage/

# Generated types (regenerate with npm run generate:types)
# src/components/types/api.ts
# src/components/types/schemas/
`
