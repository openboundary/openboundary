// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package codegen

import (
	"fmt"

	"github.com/openboundary/openboundary/internal/ir"
)

// GeneratorPlugin describes a generator with component-kind activation rules.
type GeneratorPlugin struct {
	Name         string
	NewGenerator func() Generator
	Supports     []ir.Kind // Empty means always enabled.
}

// PluginRegistry stores ordered generator plugins.
type PluginRegistry struct {
	plugins []GeneratorPlugin
	names   map[string]bool
}

// NewPluginRegistry creates an empty plugin registry.
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins: make([]GeneratorPlugin, 0),
		names:   make(map[string]bool),
	}
}

// Register adds a plugin to the registry, preserving insertion order.
func (r *PluginRegistry) Register(plugin GeneratorPlugin) error {
	if plugin.Name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	if plugin.NewGenerator == nil {
		return fmt.Errorf("plugin %q has nil generator constructor", plugin.Name)
	}
	if r.names[plugin.Name] {
		return fmt.Errorf("plugin %q already registered", plugin.Name)
	}

	r.plugins = append(r.plugins, plugin)
	r.names[plugin.Name] = true
	return nil
}

// GeneratorsForIR returns generators enabled for the provided IR.
func (r *PluginRegistry) GeneratorsForIR(i *ir.IR) ([]Generator, error) {
	generators := make([]Generator, 0, len(r.plugins))

	for _, plugin := range r.plugins {
		if !pluginEnabledForIR(plugin, i) {
			continue
		}
		generators = append(generators, plugin.NewGenerator())
	}

	return generators, nil
}

func pluginEnabledForIR(plugin GeneratorPlugin, i *ir.IR) bool {
	if len(plugin.Supports) == 0 {
		return true
	}
	if i == nil {
		return false
	}

	for _, comp := range i.Components {
		for _, kind := range plugin.Supports {
			if comp.Kind == kind {
				return true
			}
		}
	}

	return false
}
