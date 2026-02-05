// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package typescript

import (
	"github.com/openboundary/openboundary/internal/codegen"
	"github.com/openboundary/openboundary/internal/ir"
)

// NewPluginRegistry returns the default TypeScript generator plugin registry.
func NewPluginRegistry() (*codegen.PluginRegistry, error) {
	registry := codegen.NewPluginRegistry()

	plugins := []codegen.GeneratorPlugin{
		{
			Name:         "typescript-project",
			NewGenerator: func() codegen.Generator { return NewProjectGenerator() },
		},
		{
			Name:         "typescript-schemas",
			NewGenerator: func() codegen.Generator { return NewSchemaGenerator() },
			Supports:     []ir.Kind{ir.KindPostgres, ir.KindMiddleware},
		},
		{
			Name:         "typescript-openapi",
			NewGenerator: func() codegen.Generator { return NewOpenAPIGenerator() },
			Supports:     []ir.Kind{ir.KindHTTPServer, ir.KindUsecase},
		},
		{
			Name:         "typescript-context",
			NewGenerator: func() codegen.Generator { return NewContextGenerator() },
			Supports:     []ir.Kind{ir.KindHTTPServer},
		},
		{
			Name:         "typescript-hono",
			NewGenerator: func() codegen.Generator { return NewHonoServerGenerator() },
			Supports:     []ir.Kind{ir.KindHTTPServer, ir.KindMiddleware, ir.KindPostgres},
		},
		{
			Name:         "typescript-usecase",
			NewGenerator: func() codegen.Generator { return NewUsecaseGenerator() },
			Supports:     []ir.Kind{ir.KindUsecase},
		},
		{
			Name:         "typescript-tests",
			NewGenerator: func() codegen.Generator { return NewTestGenerator() },
			Supports:     []ir.Kind{ir.KindHTTPServer, ir.KindMiddleware, ir.KindUsecase},
		},
		{
			Name:         "typescript-docker",
			NewGenerator: func() codegen.Generator { return NewDockerGenerator() },
			Supports:     []ir.Kind{ir.KindHTTPServer, ir.KindPostgres},
		},
		{
			Name:         "typescript-e2e",
			NewGenerator: func() codegen.Generator { return NewE2ETestGenerator() },
			Supports:     []ir.Kind{ir.KindHTTPServer},
		},
	}

	for _, plugin := range plugins {
		if err := registry.Register(plugin); err != nil {
			return nil, err
		}
	}

	return registry, nil
}
