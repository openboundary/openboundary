// Copyright 2026 OpenBoundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package typescript

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/openboundary/openboundary/internal/ir"
	"github.com/openboundary/openboundary/internal/parser"
)

func TestSchemaGenerator_Generate_FailsOnMissingSourceFile(t *testing.T) {
	i := &ir.IR{
		BaseDir: t.TempDir(),
		Spec: &parser.Spec{
			Name:    "test",
			Version: "0.0.1",
		},
		Components: map[string]*ir.Component{
			"middleware.authn": {
				ID:   "middleware.authn",
				Kind: ir.KindMiddleware,
				Middleware: &ir.MiddlewareSpec{
					Provider: "better-auth",
					Config:   "./missing-auth.config.ts",
				},
			},
		},
	}

	g := NewSchemaGenerator()
	_, err := g.Generate(i)
	if err == nil {
		t.Fatal("Generate() expected error for missing source file")
	}
}

func TestSchemaGenerator_Generate_CopiesConfiguredFiles(t *testing.T) {
	baseDir := t.TempDir()
	authConfigPath := filepath.Join(baseDir, "auth.config.ts")
	modelPath := filepath.Join(baseDir, "model.conf")
	policyPath := filepath.Join(baseDir, "policy.csv")
	pgSchemaPath := filepath.Join(baseDir, "schema.ts")

	if err := os.WriteFile(authConfigPath, []byte("export const auth = {};"), 0644); err != nil {
		t.Fatalf("write auth config: %v", err)
	}
	if err := os.WriteFile(modelPath, []byte("model"), 0644); err != nil {
		t.Fatalf("write model: %v", err)
	}
	if err := os.WriteFile(policyPath, []byte("policy"), 0644); err != nil {
		t.Fatalf("write policy: %v", err)
	}
	if err := os.WriteFile(pgSchemaPath, []byte("export const users = {};"), 0644); err != nil {
		t.Fatalf("write postgres schema: %v", err)
	}

	i := &ir.IR{
		BaseDir: baseDir,
		Spec: &parser.Spec{
			Name:    "test",
			Version: "0.0.1",
		},
		Components: map[string]*ir.Component{
			"middleware.authn": {
				ID:   "middleware.authn",
				Kind: ir.KindMiddleware,
				Middleware: &ir.MiddlewareSpec{
					Provider: "better-auth",
					Config:   "./auth.config.ts",
				},
			},
			"middleware.authz": {
				ID:   "middleware.authz",
				Kind: ir.KindMiddleware,
				Middleware: &ir.MiddlewareSpec{
					Provider: "casbin",
					Model:    "./model.conf",
					Policy:   "./policy.csv",
				},
			},
			"postgres.primary": {
				ID:   "postgres.primary",
				Kind: ir.KindPostgres,
				Postgres: &ir.PostgresSpec{
					Provider: "drizzle",
					Schema:   "./schema.ts",
				},
			},
		},
	}

	g := NewSchemaGenerator()
	output, err := g.Generate(i)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if _, ok := output.Files["src/components/middleware-authn.middleware.config.ts"]; !ok {
		t.Fatal("missing copied better-auth config")
	}
	if _, ok := output.Files["src/components/middleware-authz.middleware.model.conf"]; !ok {
		t.Fatal("missing copied casbin model")
	}
	if _, ok := output.Files["src/components/middleware-authz.middleware.policy.csv"]; !ok {
		t.Fatal("missing copied casbin policy")
	}
	if _, ok := output.Files["src/components/postgres-primary.postgres.schema.ts"]; !ok {
		t.Fatal("missing copied postgres schema")
	}
}
