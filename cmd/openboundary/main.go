// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package main provides the CLI entry point for the openboundary compiler.
package main

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

var (
	version          = "0.1.0"
	compileOutputDir string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "openboundary",
		Short: "openboundary specification compiler",
		Long: `openboundary compiles executable specifications into runnable code.

It reads YAML specification files and generates type-safe code
for various target platforms.`,
	}

	// Version flag
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("openboundary version {{.Version}}\n")

	// compile command
	compileCmd := &cobra.Command{
		Use:   "compile [spec-file]",
		Short: "Compile a specification file",
		Long:  `Compile a specification file into executable code for the target platform.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runCompile,
	}
	compileCmd.Flags().StringVarP(&compileOutputDir, "output", "o", "generated", "Output directory for generated code")

	// validate command
	validateCmd := &cobra.Command{
		Use:   "validate [spec-file]",
		Short: "Validate a specification file",
		Long:  `Validate a specification file against the openboundary schema and semantic rules.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runValidate,
	}

	rootCmd.AddCommand(compileCmd, validateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runValidate(cmd *cobra.Command, args []string) error {
	specFile := args[0]

	// Parse the specification
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

	// Build the IR and check for reference errors
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

	// Run semantic validation
	semanticErrors := typedIR.Validate()
	if len(semanticErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Semantic validation failed with %d error(s):\n", len(semanticErrors))
		for _, e := range semanticErrors {
			fmt.Fprintf(os.Stderr, "  - %s\n", e.Error())
		}
		return fmt.Errorf("semantic validation failed")
	}

	fmt.Printf("✓ %s is valid (version: %s, name: %s, %d components)\n",
		specFile, spec.Version, spec.Name, len(spec.Components))

	return nil
}

func runCompile(cmd *cobra.Command, args []string) error {
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

	// Create generators
	generators := []codegen.Generator{
		typescript.NewProjectGenerator(),
		typescript.NewSchemaGenerator(),
		typescript.NewOpenAPIGenerator(), // Generates complete OpenAPI spec for orval
		typescript.NewContextGenerator(),
		typescript.NewHonoServerGenerator(),
		typescript.NewUsecaseGenerator(),
		typescript.NewTestGenerator(),
		typescript.NewDockerGenerator(),
		typescript.NewE2ETestGenerator(),
	}

	// Collect all outputs
	allFiles := make(map[string][]byte)
	for _, gen := range generators {
		output, err := gen.Generate(typedIR)
		if err != nil {
			return fmt.Errorf("generator %s failed: %w", gen.Name(), err)
		}
		for path, content := range output.Files {
			allFiles[path] = content
		}
	}

	// Write files to output directory
	for path, content := range allFiles {
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

	fmt.Printf("\n✓ Generated %d files in %s/\n", len(allFiles), compileOutputDir)

	return nil
}
