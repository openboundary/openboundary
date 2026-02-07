// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package validator

import (
	"fmt"
	"strings"

	"github.com/openboundary/openboundary/internal/ir"
	"github.com/openboundary/openboundary/internal/openapi"
)

// IRValidator validates the IR for semantic correctness.
// Call after building the IR to check for cycles, required fields,
// cross-component constraints, etc.
type IRValidator struct{}

// NewIRValidator creates a new IR validator.
func NewIRValidator() *IRValidator {
	return &IRValidator{}
}

// Validate performs semantic validation on the IR.
func (v *IRValidator) Validate(i *ir.IR) []ValidationError {
	var errs []ValidationError

	// Check for cycles
	cycles := i.DetectCycles()
	for _, cycle := range cycles {
		errs = append(errs, ValidationError{
			Message: fmt.Sprintf("dependency cycle: %s", formatCycle(cycle)),
		})
	}

	// Validate each component
	for _, comp := range i.Components {
		compErrs := v.validateComponent(i, comp)
		errs = append(errs, compErrs...)
	}

	// Cross-component validations
	errs = append(errs, v.validateBetterAuthRequirements(i)...)

	return errs
}

func (v *IRValidator) validateComponent(i *ir.IR, comp *ir.Component) []ValidationError {
	switch comp.Kind {
	case ir.KindHTTPServer:
		return v.validateHTTPServer(i, comp)
	case ir.KindMiddleware:
		return v.validateMiddleware(comp)
	case ir.KindPostgres:
		return v.validatePostgres(comp)
	case ir.KindUsecase:
		return v.validateUsecase(i, comp)
	}
	return nil
}

func (v *IRValidator) validateHTTPServer(i *ir.IR, comp *ir.Component) []ValidationError {
	var errs []ValidationError
	s := comp.HTTPServer

	if s == nil {
		return []ValidationError{{ID: comp.ID, Message: "missing http.server spec"}}
	}

	if s.Framework == "" {
		errs = append(errs, ValidationError{ID: comp.ID, Message: "missing required field: framework"})
	}
	if s.Port == 0 {
		errs = append(errs, ValidationError{ID: comp.ID, Message: "missing required field: port"})
	}
	if s.Port < 0 || s.Port > 65535 {
		errs = append(errs, ValidationError{ID: comp.ID, Message: "port must be between 1 and 65535"})
	}

	// Validate middleware references point to middleware components
	for _, ref := range s.Middleware {
		if sym, ok := i.Symbols.Lookup(ref); ok {
			if sym.Kind != ir.KindMiddleware {
				errs = append(errs, ValidationError{
					ID:      comp.ID,
					Message: fmt.Sprintf("middleware reference %q points to %s, expected middleware", ref, sym.Kind),
				})
			}
		}
	}

	return errs
}

func (v *IRValidator) validateMiddleware(comp *ir.Component) []ValidationError {
	var errs []ValidationError
	s := comp.Middleware

	if s == nil {
		return []ValidationError{{ID: comp.ID, Message: "missing middleware spec"}}
	}

	if s.Provider == "" {
		errs = append(errs, ValidationError{ID: comp.ID, Message: "missing required field: provider"})
	}

	// Provider-specific validation
	switch s.Provider {
	case "better-auth":
		if s.Config == "" {
			errs = append(errs, ValidationError{ID: comp.ID, Message: "better-auth provider requires config field"})
		}
	case "casbin":
		if s.Model == "" {
			errs = append(errs, ValidationError{ID: comp.ID, Message: "casbin provider requires model field"})
		}
		if s.Policy == "" {
			errs = append(errs, ValidationError{ID: comp.ID, Message: "casbin provider requires policy field"})
		}
	}

	return errs
}

func (v *IRValidator) validatePostgres(comp *ir.Component) []ValidationError {
	var errs []ValidationError
	s := comp.Postgres

	if s == nil {
		return []ValidationError{{ID: comp.ID, Message: "missing postgres spec"}}
	}

	if s.Provider == "" {
		errs = append(errs, ValidationError{ID: comp.ID, Message: "missing required field: provider"})
	}
	if s.Schema == "" {
		errs = append(errs, ValidationError{ID: comp.ID, Message: "missing required field: schema"})
	}

	return errs
}

func (v *IRValidator) validateUsecase(i *ir.IR, comp *ir.Component) []ValidationError {
	var errs []ValidationError
	s := comp.Usecase

	if s == nil {
		return []ValidationError{{ID: comp.ID, Message: "missing usecase spec"}}
	}

	if s.BindsTo == "" {
		errs = append(errs, ValidationError{ID: comp.ID, Message: "missing required field: binds_to"})
	} else {
		// Use the canonical ParseBinding from the openapi package
		serverID, _, _, err := openapi.ParseBinding(s.BindsTo)
		if err != nil {
			errs = append(errs, ValidationError{ID: comp.ID, Message: err.Error()})
		}

		// Validate the server reference exists and is an http.server
		if serverID != "" {
			if sym, ok := i.Symbols.Lookup(serverID); ok {
				if sym.Kind != ir.KindHTTPServer {
					errs = append(errs, ValidationError{
						ID:      comp.ID,
						Message: fmt.Sprintf("binds_to references %q which is %s, expected http.server", serverID, sym.Kind),
					})
				}
			}
		}
	}

	if s.Goal == "" {
		errs = append(errs, ValidationError{ID: comp.ID, Message: "missing required field: goal"})
	}

	// Validate middleware references
	for _, ref := range s.Middleware {
		if sym, ok := i.Symbols.Lookup(ref); ok {
			if sym.Kind != ir.KindMiddleware {
				errs = append(errs, ValidationError{
					ID:      comp.ID,
					Message: fmt.Sprintf("middleware reference %q points to %s, expected middleware", ref, sym.Kind),
				})
			}
		}
	}

	return errs
}

func (v *IRValidator) validateBetterAuthRequirements(i *ir.IR) []ValidationError {
	var betterAuthIDs []string
	for _, comp := range i.Components {
		if comp.Kind == ir.KindMiddleware && comp.Middleware != nil && comp.Middleware.Provider == "better-auth" {
			betterAuthIDs = append(betterAuthIDs, comp.ID)
		}
	}
	if len(betterAuthIDs) == 0 {
		return nil
	}

	betterAuthSet := make(map[string]bool, len(betterAuthIDs))
	for _, id := range betterAuthIDs {
		betterAuthSet[id] = true
	}

	// Only enforce if better-auth is actually required by a server or usecase
	required := false
	for _, comp := range i.Components {
		switch comp.Kind {
		case ir.KindHTTPServer:
			if comp.HTTPServer != nil {
				for _, ref := range comp.HTTPServer.Middleware {
					if betterAuthSet[ref] {
						required = true
						break
					}
				}
			}
		case ir.KindUsecase:
			if comp.Usecase != nil {
				for _, ref := range comp.Usecase.Middleware {
					if betterAuthSet[ref] {
						required = true
						break
					}
				}
			}
		}
		if required {
			break
		}
	}
	if !required {
		return nil
	}

	hasServer := false
	hasDrizzle := false
	for _, comp := range i.Components {
		switch comp.Kind {
		case ir.KindHTTPServer:
			if comp.HTTPServer != nil {
				hasServer = true
			}
		case ir.KindPostgres:
			if comp.Postgres != nil && comp.Postgres.Provider == "drizzle" {
				hasDrizzle = true
			}
		}
	}

	var errs []ValidationError
	if !hasServer {
		errs = append(errs, ValidationError{
			Message: "better-auth middleware requires at least one http.server component",
		})
	}
	if !hasDrizzle {
		errs = append(errs, ValidationError{
			Message: "better-auth middleware requires a postgres component with provider \"drizzle\"",
		})
	}

	return errs
}

func formatCycle(cycle []string) string {
	if len(cycle) == 0 {
		return ""
	}
	return strings.Join(cycle, " -> ") + " -> " + cycle[0]
}
