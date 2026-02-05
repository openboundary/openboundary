// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package typescript

import (
	"sort"
	"strings"

	"github.com/openboundary/openboundary/internal/ir"
)

func getUsecasesBoundToServer(i *ir.IR, serverID string) []*ir.Component {
	var usecases []*ir.Component
	if i == nil {
		return usecases
	}

	for _, comp := range i.Components {
		if comp.Kind != ir.KindUsecase || comp.Usecase == nil {
			continue
		}
		if comp.Usecase.Binding != nil && comp.Usecase.Binding.ServerID == serverID {
			usecases = append(usecases, comp)
		}
	}

	// Sort for deterministic output
	sort.Slice(usecases, func(i, j int) bool {
		return usecases[i].ID < usecases[j].ID
	})

	return usecases
}

func effectiveUsecaseMiddleware(uc *ir.Component, server *ir.Component) []string {
	if uc == nil || uc.Usecase == nil {
		return nil
	}
	// Nil means "not specified" - default to server middleware
	if uc.Usecase.Middleware == nil {
		if server != nil && server.HTTPServer != nil {
			return server.HTTPServer.Middleware
		}
		return nil
	}
	return uc.Usecase.Middleware
}

func collectServerMiddleware(i *ir.IR, server *ir.Component) []string {
	if server == nil || server.HTTPServer == nil {
		return nil
	}

	seen := make(map[string]bool)
	var ordered []string

	// Start with server defaults in order
	for _, mw := range server.HTTPServer.Middleware {
		if mw == "" || seen[mw] {
			continue
		}
		seen[mw] = true
		ordered = append(ordered, mw)
	}

	// Add middleware referenced by usecases (preserve deterministic order)
	for _, uc := range getUsecasesBoundToServer(i, server.ID) {
		for _, mw := range effectiveUsecaseMiddleware(uc, server) {
			if mw == "" || seen[mw] {
				continue
			}
			seen[mw] = true
			ordered = append(ordered, mw)
		}
	}

	return ordered
}

func serverHasPostgres(i *ir.IR, server *ir.Component) bool {
	if server == nil {
		return false
	}
	for _, dep := range server.Dependencies {
		if dep.Kind == ir.KindPostgres {
			return true
		}
	}
	if server.HTTPServer != nil && i != nil {
		for _, depID := range server.HTTPServer.DependsOn {
			if dep, ok := i.Components[depID]; ok && dep.Kind == ir.KindPostgres {
				return true
			}
		}
	}
	return false
}

func getServerPostgresDependencies(i *ir.IR, server *ir.Component) []*ir.Component {
	var deps []*ir.Component
	if server == nil {
		return deps
	}

	for _, dep := range server.Dependencies {
		if dep.Kind == ir.KindPostgres {
			deps = append(deps, dep)
		}
	}
	if len(deps) > 0 || server.HTTPServer == nil || i == nil {
		return deps
	}
	for _, depID := range server.HTTPServer.DependsOn {
		if dep, ok := i.Components[depID]; ok && dep.Kind == ir.KindPostgres {
			deps = append(deps, dep)
		}
	}
	return deps
}

func middlewareContextKeys(i *ir.IR, mwID string) []string {
	if mwID == "" {
		return nil
	}
	if i != nil {
		if comp, ok := i.Components[mwID]; ok && comp.Middleware != nil {
			switch comp.Middleware.Provider {
			case "better-auth":
				return []string{"auth"}
			case "casbin":
				return []string{"enforcer"}
			}
		}
	}

	keys := []string{}
	if strings.Contains(mwID, "authn") {
		keys = append(keys, "auth")
	}
	if strings.Contains(mwID, "authz") {
		keys = append(keys, "enforcer")
	}
	return keys
}

func contextFieldsForUsecase(i *ir.IR, uc *ir.Component, server *ir.Component) []string {
	hasDB := serverHasPostgres(i, server)
	hasAuth := false
	hasEnforcer := false

	for _, mwID := range effectiveUsecaseMiddleware(uc, server) {
		for _, key := range middlewareContextKeys(i, mwID) {
			switch key {
			case "auth":
				hasAuth = true
			case "enforcer":
				hasEnforcer = true
			}
		}
	}

	var fields []string
	if hasDB {
		fields = append(fields, "db")
	}
	if hasAuth {
		fields = append(fields, "auth")
	}
	if hasEnforcer {
		fields = append(fields, "enforcer")
	}
	return fields
}
