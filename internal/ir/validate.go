// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package ir

import (
	"fmt"
	"strings"
)

// ValidationError represents a validation error with location info.
type ValidationError struct {
	Message  string
	ID       string
	Position Position
}

// Position is a simplified position for validation errors.
type Position struct {
	File   string
	Line   int
	Column int
}

func (e *ValidationError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("%s: %s", e.ID, e.Message)
	}
	return e.Message
}

// Validate performs semantic validation on the IR.
// Call this after Build() to check for semantic errors.
//
// needs study
func (ir *IR) Validate() []error {
	var errs []error

	// Check for cycles
	cycles := ir.DetectCycles()
	for _, cycle := range cycles {
		errs = append(errs, &ValidationError{
			Message: fmt.Sprintf("dependency cycle: %s", formatCycle(cycle)),
		})
	}

	// Validate each component
	for _, comp := range ir.Components {
		compErrs := ir.validateComponent(comp)
		errs = append(errs, compErrs...)
	}

	// Cross-component validations
	errs = append(errs, ir.validateBetterAuthRequirements()...)

	return errs
}

func (ir *IR) validateComponent(comp *Component) []error {
	var errs []error

	switch comp.Kind {
	case KindHTTPServer:
		errs = ir.validateHTTPServer(comp)
	case KindMiddleware:
		errs = ir.validateMiddleware(comp)
	case KindPostgres:
		errs = ir.validatePostgres(comp)
	case KindUsecase:
		errs = ir.validateUsecase(comp)
	}

	return errs
}

func (ir *IR) validateHTTPServer(comp *Component) []error {
	var errs []error
	s := comp.HTTPServer

	if s == nil {
		return []error{&ValidationError{ID: comp.ID, Message: "missing http.server spec"}}
	}

	if s.Framework == "" {
		errs = append(errs, &ValidationError{ID: comp.ID, Message: "missing required field: framework"})
	}
	if s.Port == 0 {
		errs = append(errs, &ValidationError{ID: comp.ID, Message: "missing required field: port"})
	}
	if s.Port < 0 || s.Port > 65535 {
		errs = append(errs, &ValidationError{ID: comp.ID, Message: "port must be between 1 and 65535"})
	}

	// Validate middleware references point to middleware components
	for _, ref := range s.Middleware {
		if sym, ok := ir.Symbols.Lookup(ref); ok {
			if sym.Kind != KindMiddleware {
				errs = append(errs, &ValidationError{
					ID:      comp.ID,
					Message: fmt.Sprintf("middleware reference %q points to %s, expected middleware", ref, sym.Kind),
				})
			}
		}
	}

	return errs
}

func (ir *IR) validateMiddleware(comp *Component) []error {
	var errs []error
	s := comp.Middleware

	if s == nil {
		return []error{&ValidationError{ID: comp.ID, Message: "missing middleware spec"}}
	}

	if s.Provider == "" {
		errs = append(errs, &ValidationError{ID: comp.ID, Message: "missing required field: provider"})
	}

	// Provider-specific validation
	switch s.Provider {
	case "better-auth":
		if s.Config == "" {
			errs = append(errs, &ValidationError{ID: comp.ID, Message: "better-auth provider requires config field"})
		}
	case "casbin":
		if s.Model == "" {
			errs = append(errs, &ValidationError{ID: comp.ID, Message: "casbin provider requires model field"})
		}
		if s.Policy == "" {
			errs = append(errs, &ValidationError{ID: comp.ID, Message: "casbin provider requires policy field"})
		}
	}

	return errs
}

func (ir *IR) validatePostgres(comp *Component) []error {
	var errs []error
	s := comp.Postgres

	if s == nil {
		return []error{&ValidationError{ID: comp.ID, Message: "missing postgres spec"}}
	}

	if s.Provider == "" {
		errs = append(errs, &ValidationError{ID: comp.ID, Message: "missing required field: provider"})
	}
	if s.Schema == "" {
		errs = append(errs, &ValidationError{ID: comp.ID, Message: "missing required field: schema"})
	}

	return errs
}

func (ir *IR) validateUsecase(comp *Component) []error {
	var errs []error
	s := comp.Usecase

	if s == nil {
		return []error{&ValidationError{ID: comp.ID, Message: "missing usecase spec"}}
	}

	if s.BindsTo == "" {
		errs = append(errs, &ValidationError{ID: comp.ID, Message: "missing required field: binds_to"})
	} else {
		// Validate binds_to format: server-id:METHOD:/path
		if err := validateBindsTo(s.BindsTo); err != nil {
			errs = append(errs, &ValidationError{ID: comp.ID, Message: err.Error()})
		}

		// Validate the server reference exists and is an http.server
		serverID := extractServerFromBinding(s.BindsTo)
		if serverID != "" {
			if sym, ok := ir.Symbols.Lookup(serverID); ok {
				if sym.Kind != KindHTTPServer {
					errs = append(errs, &ValidationError{
						ID:      comp.ID,
						Message: fmt.Sprintf("binds_to references %q which is %s, expected http.server", serverID, sym.Kind),
					})
				}
			}
		}
	}

	if s.Goal == "" {
		errs = append(errs, &ValidationError{ID: comp.ID, Message: "missing required field: goal"})
	}

	// Validate middleware references
	for _, ref := range s.Middleware {
		if sym, ok := ir.Symbols.Lookup(ref); ok {
			if sym.Kind != KindMiddleware {
				errs = append(errs, &ValidationError{
					ID:      comp.ID,
					Message: fmt.Sprintf("middleware reference %q points to %s, expected middleware", ref, sym.Kind),
				})
			}
		}
	}

	return errs
}

func validateBindsTo(bindsTo string) error {
	parts := strings.SplitN(bindsTo, ":", 3)
	if len(parts) != 3 {
		return fmt.Errorf("binds_to must be in format server-id:METHOD:/path, got %q", bindsTo)
	}

	method := parts[1]
	validMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	isValid := false
	for _, m := range validMethods {
		if method == m {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid HTTP method %q in binds_to", method)
	}

	path := parts[2]
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path in binds_to must start with /, got %q", path)
	}

	return nil
}

func (ir *IR) validateBetterAuthRequirements() []error {
	var betterAuthIDs []string
	for _, comp := range ir.Components {
		if comp.Kind == KindMiddleware && comp.Middleware != nil && comp.Middleware.Provider == "better-auth" {
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
	for _, comp := range ir.Components {
		switch comp.Kind {
		case KindHTTPServer:
			if comp.HTTPServer != nil {
				for _, ref := range comp.HTTPServer.Middleware {
					if betterAuthSet[ref] {
						required = true
						break
					}
				}
			}
		case KindUsecase:
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
	for _, comp := range ir.Components {
		switch comp.Kind {
		case KindHTTPServer:
			if comp.HTTPServer != nil {
				hasServer = true
			}
		case KindPostgres:
			if comp.Postgres != nil && comp.Postgres.Provider == "drizzle" {
				hasDrizzle = true
			}
		}
	}

	var errs []error
	if !hasServer {
		errs = append(errs, &ValidationError{
			Message: "better-auth middleware requires at least one http.server component",
		})
	}
	if !hasDrizzle {
		errs = append(errs, &ValidationError{
			Message: "better-auth middleware requires a postgres component with provider \"drizzle\"",
		})
	}

	return errs
}
