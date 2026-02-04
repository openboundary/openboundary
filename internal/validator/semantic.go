package validator

import (
	"fmt"

	"github.com/openboundary/openboundary/internal/parser"
)

// SemanticValidator performs semantic validation beyond schema compliance.
type SemanticValidator struct{}

// NewSemanticValidator creates a new semantic validator.
func NewSemanticValidator() *SemanticValidator {
	return &SemanticValidator{}
}

// Validate performs semantic validation on the spec.
func (v *SemanticValidator) Validate(spec *parser.Spec) []error {
	var errs []error

	// TODO: Implement semantic validation checks:
	// - Unique component names
	// - Valid version string
	// - Required components present (e.g., at least one entry point)
	// - Port conflicts
	// - Route conflicts

	// Check for unique component names
	names := make(map[string]bool)
	for _, comp := range spec.Components {
		if names[comp.ID] {
			errs = append(errs, fmt.Errorf(
				"duplicate component name %q at %s:%d:%d",
				comp.ID, comp.Pos().File, comp.Pos().Line, comp.Pos().Column,
			))
		}
		names[comp.ID] = true
	}

	// Check version format
	if spec.Version == "" {
		errs = append(errs, fmt.Errorf("spec version is required"))
	}

	// Check spec name
	if spec.Name == "" {
		errs = append(errs, fmt.Errorf("spec name is required"))
	}

	return errs
}
