// Package validator provides validation for specification files.
package validator

import (
	"fmt"

	"github.com/openboundary/openboundary/internal/parser"
	"github.com/openboundary/openboundary/internal/schema"
)

// SchemaValidator validates components against their schemas.
type SchemaValidator struct {
	registry *schema.Registry
}

// NewSchemaValidator creates a new schema validator.
func NewSchemaValidator(registry *schema.Registry) *SchemaValidator {
	return &SchemaValidator{registry: registry}
}

// Validate validates all components in the spec against their schemas.
func (v *SchemaValidator) Validate(spec *parser.Spec) []error {
	var errs []error

	for _, comp := range spec.Components {
		kind := schema.Kind(comp.Kind)

		if !schema.IsValidKind(kind) {
			errs = append(errs, fmt.Errorf(
				"unknown component kind %q at %s:%d:%d",
				comp.Kind, comp.Pos().File, comp.Pos().Line, comp.Pos().Column,
			))
			continue
		}

		s, ok := v.registry.Get(kind)
		if !ok {
			// No schema registered for this kind, skip validation
			continue
		}

		if err := s.Validate(comp.Spec); err != nil {
			errs = append(errs, fmt.Errorf(
				"validation error for %s %q at %s:%d:%d: %w",
				comp.Kind, comp.ID,
				comp.Pos().File, comp.Pos().Line, comp.Pos().Column,
				err,
			))
		}
	}

	return errs
}
