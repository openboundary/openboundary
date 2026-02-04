package typescript

import (
	"fmt"
	"sort"
	"strings"

	"github.com/stack-bound/stack-bound/internal/codegen"
	"github.com/stack-bound/stack-bound/internal/ir"
)

// DockerGenerator generates Docker-related files for the project.
type DockerGenerator struct{}

// NewDockerGenerator creates a new Docker generator.
func NewDockerGenerator() *DockerGenerator {
	return &DockerGenerator{}
}

// Name returns the generator name.
func (g *DockerGenerator) Name() string {
	return "typescript-docker"
}

// Generate produces Docker files from the IR.
func (g *DockerGenerator) Generate(i *ir.IR) (*codegen.Output, error) {
	output := codegen.NewOutput()

	// Generate Dockerfile
	dockerfile := g.generateDockerfile()
	output.AddFile("Dockerfile", []byte(dockerfile))

	// Generate docker-compose.yml
	dockerCompose := g.generateDockerCompose(i)
	output.AddFile("docker-compose.yml", []byte(dockerCompose))

	// Generate .dockerignore
	dockerignore := g.generateDockerignore()
	output.AddFile(".dockerignore", []byte(dockerignore))

	return output, nil
}

func (g *DockerGenerator) generateDockerfile() string {
	var sb strings.Builder

	sb.WriteString(`# syntax=docker/dockerfile:1

# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Production stage
FROM node:20-alpine AS production

WORKDIR /app

# Install production dependencies only
COPY package*.json ./
RUN npm ci --omit=dev

# Copy built application from builder stage
COPY --from=builder /app/dist ./dist

# Create non-root user
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001 && \
    chown -R nodejs:nodejs /app

USER nodejs

# Expose port (default 3000, override with PORT env var)
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD node -e "require('http').get('http://localhost:' + (process.env.PORT || 3000) + '/health', (r) => process.exit(r.statusCode === 200 ? 0 : 1))"

# Start the application
CMD ["node", "dist/index.js"]
`)

	return sb.String()
}

func (g *DockerGenerator) generateDockerCompose(i *ir.IR) string {
	var sb strings.Builder

	// Detect postgres components
	hasPostgres := false
	for _, comp := range i.Components {
		if comp.Kind == ir.KindPostgres && comp.Postgres != nil {
			hasPostgres = true
			break
		}
	}

	// Get all HTTP servers (sorted for deterministic output)
	var servers []*ir.Component
	for _, comp := range i.Components {
		if comp.Kind == ir.KindHTTPServer && comp.HTTPServer != nil {
			servers = append(servers, comp)
		}
	}
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].ID < servers[j].ID
	})

	// Determine port for first server (default 3000)
	port := 3000
	if len(servers) > 0 && servers[0].HTTPServer.Port > 0 {
		port = servers[0].HTTPServer.Port
	}

	sb.WriteString("version: '3.8'\n\n")
	sb.WriteString("services:\n")

	// Postgres service
	if hasPostgres {
		sb.WriteString("  postgres:\n")
		sb.WriteString("    image: postgres:16-alpine\n")
		sb.WriteString("    environment:\n")
		sb.WriteString("      POSTGRES_USER: ${POSTGRES_USER:-postgres}\n")
		sb.WriteString("      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-postgres}\n")
		sb.WriteString("      POSTGRES_DB: ${POSTGRES_DB:-app}\n")
		sb.WriteString("    ports:\n")
		sb.WriteString("      - \"${POSTGRES_PORT:-5432}:5432\"\n")
		sb.WriteString("    volumes:\n")
		sb.WriteString("      - postgres_data:/var/lib/postgresql/data\n")
		sb.WriteString("    healthcheck:\n")
		sb.WriteString("      test: [\"CMD-SHELL\", \"pg_isready -U ${POSTGRES_USER:-postgres}\"]\n")
		sb.WriteString("      interval: 10s\n")
		sb.WriteString("      timeout: 5s\n")
		sb.WriteString("      retries: 5\n")
		sb.WriteString("    networks:\n")
		sb.WriteString("      - app_network\n\n")
	}

	// App service
	sb.WriteString("  app:\n")
	sb.WriteString("    build:\n")
	sb.WriteString("      context: .\n")
	sb.WriteString("      dockerfile: Dockerfile\n")
	sb.WriteString("      target: production\n")
	sb.WriteString(fmt.Sprintf("    ports:\n      - \"${PORT:-%d}:%d\"\n", port, port))
	sb.WriteString("    environment:\n")
	sb.WriteString(fmt.Sprintf("      PORT: ${PORT:-%d}\n", port))
	sb.WriteString("      NODE_ENV: ${NODE_ENV:-production}\n")

	if hasPostgres {
		// Construct DATABASE_URL
		sb.WriteString("      DATABASE_URL: postgres://${POSTGRES_USER:-postgres}:${POSTGRES_PASSWORD:-postgres}@postgres:5432/${POSTGRES_DB:-app}\n")
		sb.WriteString("    depends_on:\n")
		sb.WriteString("      postgres:\n")
		sb.WriteString("        condition: service_healthy\n")
	}

	sb.WriteString("    networks:\n")
	sb.WriteString("      - app_network\n")
	sb.WriteString("    restart: unless-stopped\n")

	// Networks
	sb.WriteString("\nnetworks:\n")
	sb.WriteString("  app_network:\n")
	sb.WriteString("    driver: bridge\n")

	// Volumes
	if hasPostgres {
		sb.WriteString("\nvolumes:\n")
		sb.WriteString("  postgres_data:\n")
	}

	return sb.String()
}

func (g *DockerGenerator) generateDockerignore() string {
	return `# Dependencies
node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*

# Build output
dist/
build/

# Environment files
.env
.env.local
.env.*.local

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Git
.git/
.gitignore

# Test coverage
coverage/

# Logs
logs/
*.log

# Docker
Dockerfile
docker-compose.yml
.dockerignore

# Documentation
*.md
docs/

# CI/CD
.github/
.gitlab-ci.yml
.travis.yml
`
}
