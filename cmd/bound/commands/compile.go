package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openboundary/openboundary/internal/codegen"
	"github.com/openboundary/openboundary/internal/codegen/typescript"
	"github.com/openboundary/openboundary/internal/ir"
	"github.com/openboundary/openboundary/internal/parser"
	"github.com/openboundary/openboundary/internal/validator"
	"github.com/spf13/cobra"
)

// TODO: This could be moved to dedicated pipeline package
func Compile(cmd *cobra.Command, args []string) error {
	specFile := args[0]

	// Parse and validate
	p := parser.NewParser(specFile)
	spec, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Create JSON Schema validator
	jsValidator, err := validator.NewJSONSchemaValidator()
	if err != nil {
		return fmt.Errorf("failed to initialize schema validator: %w", err)
	}

	// Validate against JSON Schema
	schemaErrors := jsValidator.Validate(spec)
	if len(schemaErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Schema validation failed with %d error(s):\n", len(schemaErrors))
		for _, e := range schemaErrors {
			fmt.Fprintf(os.Stderr, "  - %s\n", e.Error())
		}
		return fmt.Errorf("schema validation failed")
	}

	// Build the IR
	baseDir := filepath.Dir(specFile)
	builder := ir.NewBuilder().WithBaseDir(baseDir)
	typedIR, buildErrors := builder.Build(spec)
	if len(buildErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Build failed with %d error(s):\n", len(buildErrors))
		for _, e := range buildErrors {
			fmt.Fprintf(os.Stderr, "  - %s\n", e.Error())
		}
		return fmt.Errorf("build failed")
	}

	// Validate
	if errs := typedIR.Validate(); len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "Validation failed with %d error(s):\n", len(errs))
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  - %s\n", e.Error())
		}
		return fmt.Errorf("validation failed")
	}

	fmt.Printf("Compiling specification: %s (%d components)\n", specFile, len(typedIR.Components))

	pluginRegistry, err := typescript.NewPlugin Registry()
	if err != nil {
		return fmt.Errorf("failed to initialize TypeScript plugin registry: %w", err)
	}

	generators, err := pluginRegistry.GeneratorsForIR(typedIR)
	if err != nil {
		return fmt.Errorf("failed to resolve TypeScript generators: %w", err)
	}

	// Plan and validate all artifacts before writing any files.
	planner := codegen.NewArtifactPlanner()
	for _, gen := range generators {
		output, genErr := gen.Generate(typedIR)
		if genErr != nil {
			return fmt.Errorf("generator %s failed: %w", gen.Name(), genErr)
		}
		if planErr := planner.AddOutput(gen.Name(), output); planErr != nil {
			return fmt.Errorf("artifact planning failed for %s: %w", gen.Name(), planErr)
		}
	}

	artifacts := planner.Artifacts()

	// Write files to output directory
	for _, artifact := range artifacts {
		path := artifact.Path
		content := artifact.Content
		fullPath := filepath.Join(compileOutputDir, path)

		// Create parent directories
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		// Write file
		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", fullPath, err)
		}

		fmt.Printf("  → %s\n", path)
	}

	fmt.Printf("\n✓ Generated %d files in %s/\n", len(artifacts), compileOutputDir)

	return nil
}
