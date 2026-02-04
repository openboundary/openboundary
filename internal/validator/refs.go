package validator

import (
	"fmt"

	"github.com/openboundary/openboundary/internal/parser"
)

// RefValidator validates cross-references between components.
type RefValidator struct{}

// NewRefValidator creates a new reference validator.
func NewRefValidator() *RefValidator {
	return &RefValidator{}
}

// Validate validates all cross-references in the spec.
func (v *RefValidator) Validate(spec *parser.Spec) []error {
	// TODO: Implement cross-reference validation
	// - Build index of all component names
	// - Find all $ref values in component specs
	// - Verify each $ref points to an existing component
	// - Verify type compatibility

	var errs []error

	// Build component index
	components := make(map[string]*parser.Component)
	for i := range spec.Components {
		comp := &spec.Components[i]
		components[comp.ID] = comp
	}

	// Find and validate references
	for _, comp := range spec.Components {
		refs := findRefs(comp.Spec)
		for _, ref := range refs {
			if _, ok := components[ref]; !ok {
				errs = append(errs, fmt.Errorf(
					"unresolved reference %q in component %q",
					ref, comp.ID,
				))
			}
		}
	}

	return errs
}

// findRefs recursively finds all $ref values in a spec.
func findRefs(spec map[string]interface{}) []string {
	var refs []string

	for key, value := range spec {
		if key == "$ref" {
			if ref, ok := value.(string); ok {
				refs = append(refs, ref)
			}
		}

		switch v := value.(type) {
		case map[string]interface{}:
			refs = append(refs, findRefs(v)...)
		case []interface{}:
			for _, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					refs = append(refs, findRefs(m)...)
				}
			}
		case string:
			// String values without $ref are not references
			// TODO: Future support for template syntax like "{{ ref:componentName }}"
		}
	}

	return refs
}
