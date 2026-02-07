// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package pipeline

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openboundary/openboundary/internal/codegen"
	"github.com/openboundary/openboundary/internal/ir"
	"github.com/openboundary/openboundary/internal/parser"
	"github.com/openboundary/openboundary/internal/validator"
)

// parseStage parses a spec file into an AST.
type parseStage struct{}

func Parse() Stage { return &parseStage{} }

func (s *parseStage) Name() string { return "parse" }

func (s *parseStage) Run(ctx *Context) error {
	p := parser.NewParser(ctx.SpecPath)
	spec, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}
	ctx.AST = spec
	return nil
}

// validateSchemaStage validates the AST against JSON Schema.
type validateSchemaStage struct{}

func ValidateSchema() Stage { return &validateSchemaStage{} }

func (s *validateSchemaStage) Name() string { return "validate-schema" }

func (s *validateSchemaStage) Run(ctx *Context) error {
	jsValidator, err := validator.NewJSONSchemaValidator()
	if err != nil {
		return fmt.Errorf("failed to initialize schema validator: %w", err)
	}

	schemaErrors := jsValidator.Validate(ctx.AST)
	if len(schemaErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Schema validation failed with %d error(s):\n", len(schemaErrors))
		for _, e := range schemaErrors {
			fmt.Fprintf(os.Stderr, "  - %s\n", e.Error())
		}
		return fmt.Errorf("schema validation failed")
	}
	return nil
}

// buildIRStage builds the IR from the AST.
type buildIRStage struct{}

func BuildIR() Stage { return &buildIRStage{} }

func (s *buildIRStage) Name() string { return "build-ir" }

func (s *buildIRStage) Run(ctx *Context) error {
	baseDir := filepath.Dir(ctx.SpecPath)
	builder := ir.NewBuilder().WithBaseDir(baseDir)
	typedIR, buildErrors := builder.Build(ctx.AST)
	if len(buildErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Build failed with %d error(s):\n", len(buildErrors))
		for _, e := range buildErrors {
			fmt.Fprintf(os.Stderr, "  - %s\n", e.Error())
		}
		return fmt.Errorf("build failed")
	}
	ctx.IR = typedIR
	return nil
}

// validateIRStage runs semantic validation on the IR.
type validateIRStage struct{}

func ValidateIR() Stage { return &validateIRStage{} }

func (s *validateIRStage) Name() string { return "validate-ir" }

func (s *validateIRStage) Run(ctx *Context) error {
	errs := ctx.IR.Validate()
	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "Validation failed with %d error(s):\n", len(errs))
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  - %s\n", e.Error())
		}
		return fmt.Errorf("validation failed")
	}
	return nil
}

// generateStage resolves generators from a plugin registry and produces artifacts.
type generateStage struct {
	newRegistry func() (*codegen.PluginRegistry, error)
}

func Generate(newRegistry func() (*codegen.PluginRegistry, error)) Stage {
	return &generateStage{newRegistry: newRegistry}
}

func (s *generateStage) Name() string { return "generate" }

func (s *generateStage) Run(ctx *Context) error {
	pluginRegistry, err := s.newRegistry()
	if err != nil {
		return fmt.Errorf("failed to initialize plugin registry: %w", err)
	}

	generators, err := pluginRegistry.GeneratorsForIR(ctx.IR)
	if err != nil {
		return fmt.Errorf("failed to resolve generators: %w", err)
	}

	planner := codegen.NewArtifactPlanner()
	for _, gen := range generators {
		output, genErr := gen.Generate(ctx.IR)
		if genErr != nil {
			return fmt.Errorf("generator %s failed: %w", gen.Name(), genErr)
		}
		if planErr := planner.AddOutput(gen.Name(), output); planErr != nil {
			return fmt.Errorf("artifact planning failed for %s: %w", gen.Name(), planErr)
		}
	}

	ctx.Artifacts = planner.Artifacts()
	return nil
}

// writeStage writes artifacts to the output directory.
type writeStage struct{}

func Write() Stage { return &writeStage{} }

func (s *writeStage) Name() string { return "write" }

func (s *writeStage) Run(ctx *Context) error {
	for _, artifact := range ctx.Artifacts {
		fullPath := filepath.Join(ctx.OutputDir, artifact.Path)

		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		if err := os.WriteFile(fullPath, artifact.Content, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", fullPath, err)
		}

		fmt.Printf("  → %s\n", artifact.Path)
	}
	return nil
}
