// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package cache provides compilation caching for incremental regeneration.
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/openboundary/openboundary/internal/ir"
)

const (
	// CurrentCacheVersion is the current cache format version.
	CurrentCacheVersion = "1.0"
)

// Cache represents the compilation cache structure.
type Cache struct {
	Version    string                `json:"cache_version"`
	SpecHash   string                `json:"specHash"`
	Components map[string]*Component `json:"components"`
}

// Component represents a cached component with its hash and artifacts.
type Component struct {
	Hash      string   `json:"hash"`
	Artifacts []string `json:"artifacts"`
}

// New creates a new empty cache with the current version.
func New() *Cache {
	return &Cache{
		Version:    CurrentCacheVersion,
		Components: make(map[string]*Component),
	}
}

// Load loads the cache from the given path.
// Returns an empty cache if the file doesn't exist or has incompatible version.
func Load(path string) (*Cache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return New(), nil
		}
		return nil, fmt.Errorf("failed to read cache: %w", err)
	}

	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to parse cache: %w", err)
	}

	// Validate and migrate cache version
	migrated, err := migrateCache(&cache)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate cache: %w", err)
	}

	if migrated.Components == nil {
		migrated.Components = make(map[string]*Component)
	}

	return migrated, nil
}

// migrateCache validates the cache version and performs any necessary migrations.
// Returns a new cache if the version is incompatible or too old to migrate.
func migrateCache(cache *Cache) (*Cache, error) {
	// If no version field exists, this is a legacy cache (pre-1.0)
	if cache.Version == "" {
		// For now, we invalidate legacy caches and start fresh
		// Future versions can add migration logic here
		return New(), nil
	}

	// If version matches current, no migration needed
	if cache.Version == CurrentCacheVersion {
		return cache, nil
	}

	// Handle future version migrations here
	// For example:
	// if cache.Version == "1.0" {
	//     return migrateFrom1_0To1_1(cache)
	// }

	// Unknown version - start with fresh cache
	return New(), nil
}

// Save saves the cache to the given path.
func (c *Cache) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	return nil
}

// ComputeSpecHash computes a semantic hash of the entire specification.
// This hash ignores comments, whitespace, and position information,
// focusing only on semantically meaningful content.
func ComputeSpecHash(irData *ir.IR) (string, error) {
	// Create a canonical representation of the spec
	// Only include semantically meaningful fields
	canonical := map[string]interface{}{
		"version":     irData.Spec.Version,
		"name":        irData.Spec.Name,
		"description": irData.Spec.Description,
	}

	// Add component IDs and kinds in sorted order
	// We don't include full component specs here as those are hashed separately
	components := make([]map[string]string, 0, len(irData.Spec.Components))
	for _, comp := range irData.Spec.Components {
		components = append(components, map[string]string{
			"id":   comp.ID,
			"kind": comp.Kind,
		})
	}

	// Sort components by ID for deterministic ordering
	sort.Slice(components, func(i, j int) bool {
		return components[i]["id"] < components[j]["id"]
	})
	canonical["components"] = components

	// Marshal to JSON with sorted keys for deterministic hashing
	data, err := json.Marshal(canonical)
	if err != nil {
		return "", fmt.Errorf("failed to marshal spec: %w", err)
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// ComputeComponentHash computes a semantic hash of a single component.
// This hash ignores comments, whitespace, and field order,
// focusing only on semantically meaningful content.
func ComputeComponentHash(comp *ir.Component) (string, error) {
	// Create a canonical representation of the component
	// Only include semantically meaningful fields
	data := map[string]interface{}{
		"id":   comp.ID,
		"kind": string(comp.Kind),
	}

	// Add kind-specific fields with sorted arrays for deterministic hashing
	switch comp.Kind {
	case ir.KindHTTPServer:
		if comp.HTTPServer != nil {
			data["spec"] = map[string]interface{}{
				"framework":  comp.HTTPServer.Framework,
				"port":       comp.HTTPServer.Port,
				"openapi":    comp.HTTPServer.OpenAPI,
				"middleware": sortedCopy(comp.HTTPServer.Middleware),
				"depends_on": sortedCopy(comp.HTTPServer.DependsOn),
			}
		}
	case ir.KindMiddleware:
		if comp.Middleware != nil {
			data["spec"] = map[string]interface{}{
				"provider":   comp.Middleware.Provider,
				"config":     comp.Middleware.Config,
				"model":      comp.Middleware.Model,
				"policy":     comp.Middleware.Policy,
				"depends_on": sortedCopy(comp.Middleware.DependsOn),
			}
		}
	case ir.KindPostgres:
		if comp.Postgres != nil {
			data["spec"] = map[string]interface{}{
				"provider": comp.Postgres.Provider,
				"schema":   comp.Postgres.Schema,
			}
		}
	case ir.KindUsecase:
		if comp.Usecase != nil {
			data["spec"] = map[string]interface{}{
				"binds_to":            comp.Usecase.BindsTo,
				"middleware":          sortedCopy(comp.Usecase.Middleware),
				"goal":                comp.Usecase.Goal,
				"actor":               comp.Usecase.Actor,
				"preconditions":       sortedCopy(comp.Usecase.Preconditions),
				"acceptance_criteria": sortedCopy(comp.Usecase.AcceptanceCriteria),
				"postconditions":      sortedCopy(comp.Usecase.Postconditions),
			}
		}
	}

	// Marshal to JSON with sorted keys for deterministic hashing
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal component: %w", err)
	}

	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:]), nil
}

// sortedCopy returns a sorted copy of a string slice for deterministic hashing.
// Returns nil if the input is nil to preserve semantic meaning.
func sortedCopy(slice []string) []string {
	if slice == nil {
		return nil
	}
	if len(slice) == 0 {
		return []string{}
	}

	copied := make([]string, len(slice))
	copy(copied, slice)
	sort.Strings(copied)
	return copied
}

// Update updates the cache with new IR and component hashes.
func (c *Cache) Update(irData *ir.IR) error {
	// Update spec hash
	specHash, err := ComputeSpecHash(irData)
	if err != nil {
		return fmt.Errorf("failed to compute spec hash: %w", err)
	}
	c.SpecHash = specHash

	// Update component hashes
	c.Components = make(map[string]*Component)
	for id, comp := range irData.Components {
		hash, err := ComputeComponentHash(comp)
		if err != nil {
			return fmt.Errorf("failed to compute hash for component %s: %w", id, err)
		}

		c.Components[id] = &Component{
			Hash:      hash,
			Artifacts: []string{}, // Will be populated by artifact tracking
		}
	}

	return nil
}

// SetArtifacts sets the artifacts for a component.
func (c *Cache) SetArtifacts(componentID string, artifacts []string) {
	if comp, exists := c.Components[componentID]; exists {
		comp.Artifacts = artifacts
	}
}

// GetComponentHash returns the hash for a component, or empty string if not found.
func (c *Cache) GetComponentHash(componentID string) string {
	if comp, exists := c.Components[componentID]; exists {
		return comp.Hash
	}
	return ""
}

// HasComponent returns true if the cache contains the given component.
func (c *Cache) HasComponent(componentID string) bool {
	_, exists := c.Components[componentID]
	return exists
}
