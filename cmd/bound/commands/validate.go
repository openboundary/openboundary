package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openboundary/openboundary/internal/ir"
	"github.com/openboundary/openboundary/internal/parser"
	"github.com/openboundary/openboundary/internal/validator"
	"github.com/spf13/cobra"
)

func Validate(cmd *cobra.Command, args []string) error {
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
