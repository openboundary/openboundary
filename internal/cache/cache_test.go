// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/openboundary/openboundary/internal/ir"
	"github.com/openboundary/openboundary/internal/parser"
)

func TestNew(t *testing.T) {
	c := New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
	if c.Version != CurrentCacheVersion {
		t.Errorf("Expected version %s, got %s", CurrentCacheVersion, c.Version)
	}
	if c.Components == nil {
		t.Fatal("Components map is nil")
	}
	if len(c.Components) != 0 {
		t.Errorf("Expected empty components, got %d", len(c.Components))
	}
}

func TestLoadNonExistent(t *testing.T) {
	c, err := Load("/nonexistent/cache.json")
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if c == nil {
		t.Fatal("Load() returned nil cache")
	}
	if len(c.Components) != 0 {
		t.Errorf("Expected empty cache, got %d components", len(c.Components))
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, ".bound", "cache.json")

	// Create a cache with version
	c := New()
	c.SpecHash = "abc123"
	c.Components["test.component"] = &Component{
		Hash:      "def456",
		Artifacts: []string{"src/test.ts"},
	}

	// Save it
	if err := c.Save(cachePath); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Load it back
	loaded, err := Load(cachePath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify
	if loaded.Version != CurrentCacheVersion {
		t.Errorf("Version mismatch: got %s, want %s", loaded.Version, CurrentCacheVersion)
	}
	if loaded.SpecHash != c.SpecHash {
		t.Errorf("SpecHash mismatch: got %s, want %s", loaded.SpecHash, c.SpecHash)
	}
	if len(loaded.Components) != len(c.Components) {
		t.Errorf("Components count mismatch: got %d, want %d", len(loaded.Components), len(c.Components))
	}
	if comp, ok := loaded.Components["test.component"]; ok {
		if comp.Hash != "def456" {
			t.Errorf("Component hash mismatch: got %s, want def456", comp.Hash)
		}
		if len(comp.Artifacts) != 1 || comp.Artifacts[0] != "src/test.ts" {
			t.Errorf("Artifacts mismatch: got %v, want [src/test.ts]", comp.Artifacts)
		}
	} else {
		t.Error("Component not found in loaded cache")
	}
}

func TestComputeSpecHash(t *testing.T) {
	spec := &parser.Spec{
		Version: "0.1",
		Name:    "test",
		Components: []parser.Component{
			{ID: "test.server", Kind: "http.server"},
		},
	}

	irData := ir.New(spec)

	hash1, err := ComputeSpecHash(irData)
	if err != nil {
		t.Fatalf("ComputeSpecHash() failed: %v", err)
	}
	if hash1 == "" {
		t.Error("Hash is empty")
	}

	// Same spec should produce same hash
	hash2, err := ComputeSpecHash(irData)
	if err != nil {
		t.Fatalf("ComputeSpecHash() failed: %v", err)
	}
	if hash1 != hash2 {
		t.Errorf("Hash not deterministic: %s != %s", hash1, hash2)
	}

	// Different spec should produce different hash
	spec.Name = "different"
	hash3, err := ComputeSpecHash(irData)
	if err != nil {
		t.Fatalf("ComputeSpecHash() failed: %v", err)
	}
	if hash1 == hash3 {
		t.Error("Different specs produced same hash")
	}
}

func TestComputeComponentHash(t *testing.T) {
	tests := []struct {
		name      string
		component *ir.Component
		wantErr   bool
	}{
		{
			name: "http.server component",
			component: &ir.Component{
				ID:   "test.server",
				Kind: ir.KindHTTPServer,
				HTTPServer: &ir.HTTPServerSpec{
					Framework:  "express",
					Port:       3000,
					OpenAPI:    "api.yaml",
					Middleware: []string{"auth"},
					DependsOn:  []string{"db"},
				},
			},
			wantErr: false,
		},
		{
			name: "middleware component",
			component: &ir.Component{
				ID:   "test.auth",
				Kind: ir.KindMiddleware,
				Middleware: &ir.MiddlewareSpec{
					Provider: "clerk",
					Config:   "config.yaml",
				},
			},
			wantErr: false,
		},
		{
			name: "postgres component",
			component: &ir.Component{
				ID:   "test.db",
				Kind: ir.KindPostgres,
				Postgres: &ir.PostgresSpec{
					Provider: "neon",
					Schema:   "schema.sql",
				},
			},
			wantErr: false,
		},
		{
			name: "usecase component",
			component: &ir.Component{
				ID:   "test.usecase",
				Kind: ir.KindUsecase,
				Usecase: &ir.UsecaseSpec{
					BindsTo:            "GET /users",
					Goal:               "List users",
					Actor:              "Admin",
					Preconditions:      []string{"Authenticated"},
					AcceptanceCriteria: []string{"Returns user list"},
					Postconditions:     []string{"None"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := ComputeComponentHash(tt.component)
			if (err != nil) != tt.wantErr {
				t.Errorf("ComputeComponentHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && hash == "" {
				t.Error("Hash is empty")
			}

			// Verify determinism
			hash2, err := ComputeComponentHash(tt.component)
			if err != nil {
				t.Fatalf("Second hash computation failed: %v", err)
			}
			if hash != hash2 {
				t.Errorf("Hash not deterministic: %s != %s", hash, hash2)
			}
		})
	}
}

func TestComputeComponentHashChanges(t *testing.T) {
	comp := &ir.Component{
		ID:   "test.server",
		Kind: ir.KindHTTPServer,
		HTTPServer: &ir.HTTPServerSpec{
			Framework: "express",
			Port:      3000,
		},
	}

	hash1, err := ComputeComponentHash(comp)
	if err != nil {
		t.Fatalf("ComputeComponentHash() failed: %v", err)
	}

	// Modify component
	comp.HTTPServer.Port = 4000

	hash2, err := ComputeComponentHash(comp)
	if err != nil {
		t.Fatalf("ComputeComponentHash() failed: %v", err)
	}

	if hash1 == hash2 {
		t.Error("Modified component produced same hash")
	}
}

func TestUpdate(t *testing.T) {
	spec := &parser.Spec{
		Version: "0.1",
		Name:    "test",
	}
	irData := ir.New(spec)
	irData.Components["test.server"] = &ir.Component{
		ID:   "test.server",
		Kind: ir.KindHTTPServer,
		HTTPServer: &ir.HTTPServerSpec{
			Framework: "express",
			Port:      3000,
		},
	}

	cache := New()
	err := cache.Update(irData)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	// Verify spec hash was computed
	if cache.SpecHash == "" {
		t.Error("SpecHash is empty")
	}

	// Verify component was cached
	if len(cache.Components) != 1 {
		t.Errorf("Expected 1 component, got %d", len(cache.Components))
	}

	if comp, ok := cache.Components["test.server"]; ok {
		if comp.Hash == "" {
			t.Error("Component hash is empty")
		}
	} else {
		t.Error("Component not found in cache")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache.json")

	// Write invalid JSON
	if err := os.WriteFile(cachePath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := Load(cachePath)
	if err == nil {
		t.Error("Expected error loading invalid JSON, got nil")
	}
}

func TestCacheVersioning(t *testing.T) {
	t.Run("new cache has current version", func(t *testing.T) {
		c := New()
		if c.Version != CurrentCacheVersion {
			t.Errorf("Expected version %s, got %s", CurrentCacheVersion, c.Version)
		}
	})

	t.Run("saved cache includes version", func(t *testing.T) {
		tmpDir := t.TempDir()
		cachePath := filepath.Join(tmpDir, "cache.json")

		c := New()
		c.SpecHash = "test123"

		if err := c.Save(cachePath); err != nil {
			t.Fatalf("Save() failed: %v", err)
		}

		// Read raw JSON to verify version is present
		data, err := os.ReadFile(cachePath)
		if err != nil {
			t.Fatalf("Failed to read cache file: %v", err)
		}

		var raw map[string]interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}

		version, ok := raw["cache_version"].(string)
		if !ok {
			t.Error("cache_version field missing or not a string")
		}
		if version != CurrentCacheVersion {
			t.Errorf("Expected version %s in JSON, got %s", CurrentCacheVersion, version)
		}
	})
}

func TestCacheMigration(t *testing.T) {
	t.Run("legacy cache without version returns empty cache", func(t *testing.T) {
		tmpDir := t.TempDir()
		cachePath := filepath.Join(tmpDir, "cache.json")

		// Create legacy cache (no version field)
		legacyCache := map[string]interface{}{
			"specHash": "legacy123",
			"components": map[string]interface{}{
				"test.comp": map[string]interface{}{
					"hash":      "hash123",
					"artifacts": []string{"file.ts"},
				},
			},
		}

		data, err := json.MarshalIndent(legacyCache, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal legacy cache: %v", err)
		}

		if err := os.WriteFile(cachePath, data, 0644); err != nil {
			t.Fatalf("Failed to write legacy cache: %v", err)
		}

		// Load should migrate to empty cache
		loaded, err := Load(cachePath)
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if loaded.Version != CurrentCacheVersion {
			t.Errorf("Expected migrated version %s, got %s", CurrentCacheVersion, loaded.Version)
		}

		// Legacy cache should be invalidated (empty)
		if loaded.SpecHash != "" {
			t.Error("Expected empty SpecHash after migration from legacy")
		}
		if len(loaded.Components) != 0 {
			t.Errorf("Expected empty components after migration from legacy, got %d", len(loaded.Components))
		}
	})

	t.Run("current version cache loads normally", func(t *testing.T) {
		tmpDir := t.TempDir()
		cachePath := filepath.Join(tmpDir, "cache.json")

		// Create cache with current version
		c := New()
		c.SpecHash = "current123"
		c.Components["test.comp"] = &Component{
			Hash:      "hash456",
			Artifacts: []string{"src/test.ts"},
		}

		if err := c.Save(cachePath); err != nil {
			t.Fatalf("Save() failed: %v", err)
		}

		// Load should work without migration
		loaded, err := Load(cachePath)
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if loaded.Version != CurrentCacheVersion {
			t.Errorf("Expected version %s, got %s", CurrentCacheVersion, loaded.Version)
		}

		if loaded.SpecHash != c.SpecHash {
			t.Errorf("Expected SpecHash %s, got %s", c.SpecHash, loaded.SpecHash)
		}

		if len(loaded.Components) != 1 {
			t.Errorf("Expected 1 component, got %d", len(loaded.Components))
		}
	})

	t.Run("unknown future version returns empty cache", func(t *testing.T) {
		tmpDir := t.TempDir()
		cachePath := filepath.Join(tmpDir, "cache.json")

		// Create cache with future version
		futureCache := map[string]interface{}{
			"cache_version": "99.0",
			"specHash":      "future123",
			"components": map[string]interface{}{
				"test.comp": map[string]interface{}{
					"hash":      "hash789",
					"artifacts": []string{"src/future.ts"},
				},
			},
		}

		data, err := json.MarshalIndent(futureCache, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal future cache: %v", err)
		}

		if err := os.WriteFile(cachePath, data, 0644); err != nil {
			t.Fatalf("Failed to write future cache: %v", err)
		}

		// Load should invalidate unknown version
		loaded, err := Load(cachePath)
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if loaded.Version != CurrentCacheVersion {
			t.Errorf("Expected version %s after unknown version, got %s", CurrentCacheVersion, loaded.Version)
		}

		// Future cache should be invalidated (empty)
		if loaded.SpecHash != "" {
			t.Error("Expected empty SpecHash after unknown version")
		}
		if len(loaded.Components) != 0 {
			t.Errorf("Expected empty components after unknown version, got %d", len(loaded.Components))
		}
	})
}

func TestMigrateCache(t *testing.T) {
	t.Run("current version returns same cache", func(t *testing.T) {
		cache := &Cache{
			Version:  CurrentCacheVersion,
			SpecHash: "test123",
			Components: map[string]*Component{
				"comp1": {Hash: "hash1", Artifacts: []string{"file1.ts"}},
			},
		}

		migrated, err := migrateCache(cache)
		if err != nil {
			t.Fatalf("migrateCache() failed: %v", err)
		}

		if migrated != cache {
			t.Error("Expected same cache instance for current version")
		}

		if migrated.SpecHash != cache.SpecHash {
			t.Error("SpecHash should be preserved")
		}

		if len(migrated.Components) != len(cache.Components) {
			t.Error("Components should be preserved")
		}
	})

	t.Run("empty version returns new cache", func(t *testing.T) {
		cache := &Cache{
			Version:  "",
			SpecHash: "old",
			Components: map[string]*Component{
				"old": {Hash: "old", Artifacts: []string{"old.ts"}},
			},
		}

		migrated, err := migrateCache(cache)
		if err != nil {
			t.Fatalf("migrateCache() failed: %v", err)
		}

		if migrated == cache {
			t.Error("Expected new cache instance for legacy version")
		}

		if migrated.Version != CurrentCacheVersion {
			t.Errorf("Expected version %s, got %s", CurrentCacheVersion, migrated.Version)
		}

		if migrated.SpecHash != "" {
			t.Error("Expected empty SpecHash for new cache")
		}

		if len(migrated.Components) != 0 {
			t.Error("Expected empty components for new cache")
		}
	})
}

func TestSemanticSpecHashing(t *testing.T) {
	t.Run("same spec produces same hash", func(t *testing.T) {
		spec := &parser.Spec{
			Version: "0.1",
			Name:    "test-app",
			Components: []parser.Component{
				{ID: "server.api", Kind: "http.server"},
				{ID: "db.main", Kind: "postgres"},
			},
		}

		irData := ir.New(spec)
		hash1, err := ComputeSpecHash(irData)
		if err != nil {
			t.Fatalf("ComputeSpecHash() failed: %v", err)
		}

		hash2, err := ComputeSpecHash(irData)
		if err != nil {
			t.Fatalf("ComputeSpecHash() failed: %v", err)
		}

		if hash1 != hash2 {
			t.Errorf("Same spec produced different hashes: %s != %s", hash1, hash2)
		}
	})

	t.Run("component order doesn't affect hash", func(t *testing.T) {
		spec1 := &parser.Spec{
			Version: "0.1",
			Name:    "test-app",
			Components: []parser.Component{
				{ID: "server.api", Kind: "http.server"},
				{ID: "db.main", Kind: "postgres"},
			},
		}

		spec2 := &parser.Spec{
			Version: "0.1",
			Name:    "test-app",
			Components: []parser.Component{
				{ID: "db.main", Kind: "postgres"},
				{ID: "server.api", Kind: "http.server"},
			},
		}

		hash1, err := ComputeSpecHash(ir.New(spec1))
		if err != nil {
			t.Fatalf("ComputeSpecHash() failed: %v", err)
		}

		hash2, err := ComputeSpecHash(ir.New(spec2))
		if err != nil {
			t.Fatalf("ComputeSpecHash() failed: %v", err)
		}

		if hash1 != hash2 {
			t.Errorf("Different component order produced different hashes: %s != %s", hash1, hash2)
		}
	})

	t.Run("spec changes produce different hash", func(t *testing.T) {
		spec1 := &parser.Spec{
			Version: "0.1",
			Name:    "test-app",
			Components: []parser.Component{
				{ID: "server.api", Kind: "http.server"},
			},
		}

		spec2 := &parser.Spec{
			Version: "0.1",
			Name:    "test-app",
			Components: []parser.Component{
				{ID: "server.api", Kind: "http.server"},
				{ID: "db.main", Kind: "postgres"}, // Added component
			},
		}

		hash1, err := ComputeSpecHash(ir.New(spec1))
		if err != nil {
			t.Fatalf("ComputeSpecHash() failed: %v", err)
		}

		hash2, err := ComputeSpecHash(ir.New(spec2))
		if err != nil {
			t.Fatalf("ComputeSpecHash() failed: %v", err)
		}

		if hash1 == hash2 {
			t.Error("Different specs produced same hash")
		}
	})
}

func TestSemanticComponentHashing(t *testing.T) {
	t.Run("array order doesn't affect hash", func(t *testing.T) {
		comp1 := &ir.Component{
			ID:   "test.server",
			Kind: ir.KindHTTPServer,
			HTTPServer: &ir.HTTPServerSpec{
				Framework:  "express",
				Port:       3000,
				Middleware: []string{"auth", "logging", "cors"},
				DependsOn:  []string{"db", "cache"},
			},
		}

		comp2 := &ir.Component{
			ID:   "test.server",
			Kind: ir.KindHTTPServer,
			HTTPServer: &ir.HTTPServerSpec{
				Framework:  "express",
				Port:       3000,
				Middleware: []string{"cors", "logging", "auth"}, // Different order
				DependsOn:  []string{"cache", "db"},            // Different order
			},
		}

		hash1, err := ComputeComponentHash(comp1)
		if err != nil {
			t.Fatalf("ComputeComponentHash() failed: %v", err)
		}

		hash2, err := ComputeComponentHash(comp2)
		if err != nil {
			t.Fatalf("ComputeComponentHash() failed: %v", err)
		}

		if hash1 != hash2 {
			t.Errorf("Different array order produced different hashes: %s != %s", hash1, hash2)
		}
	})

	t.Run("nil vs empty array produces different hash", func(t *testing.T) {
		comp1 := &ir.Component{
			ID:   "test.server",
			Kind: ir.KindHTTPServer,
			HTTPServer: &ir.HTTPServerSpec{
				Framework:  "express",
				Port:       3000,
				Middleware: nil,
			},
		}

		comp2 := &ir.Component{
			ID:   "test.server",
			Kind: ir.KindHTTPServer,
			HTTPServer: &ir.HTTPServerSpec{
				Framework:  "express",
				Port:       3000,
				Middleware: []string{}, // Empty array instead of nil
			},
		}

		hash1, err := ComputeComponentHash(comp1)
		if err != nil {
			t.Fatalf("ComputeComponentHash() failed: %v", err)
		}

		hash2, err := ComputeComponentHash(comp2)
		if err != nil {
			t.Fatalf("ComputeComponentHash() failed: %v", err)
		}

		if hash1 == hash2 {
			t.Error("Nil vs empty array produced same hash (should differ)")
		}
	})

	t.Run("semantic changes produce different hash", func(t *testing.T) {
		comp1 := &ir.Component{
			ID:   "test.server",
			Kind: ir.KindHTTPServer,
			HTTPServer: &ir.HTTPServerSpec{
				Framework: "express",
				Port:      3000,
			},
		}

		comp2 := &ir.Component{
			ID:   "test.server",
			Kind: ir.KindHTTPServer,
			HTTPServer: &ir.HTTPServerSpec{
				Framework: "express",
				Port:      4000, // Different port
			},
		}

		hash1, err := ComputeComponentHash(comp1)
		if err != nil {
			t.Fatalf("ComputeComponentHash() failed: %v", err)
		}

		hash2, err := ComputeComponentHash(comp2)
		if err != nil {
			t.Fatalf("ComputeComponentHash() failed: %v", err)
		}

		if hash1 == hash2 {
			t.Error("Different port values produced same hash")
		}
	})
}

func TestSortedCopy(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		result := sortedCopy(nil)
		if result != nil {
			t.Errorf("Expected nil, got %v", result)
		}
	})

	t.Run("empty input returns empty", func(t *testing.T) {
		result := sortedCopy([]string{})
		if result == nil {
			t.Error("Expected empty slice, got nil")
		}
		if len(result) != 0 {
			t.Errorf("Expected empty slice, got %v", result)
		}
	})

	t.Run("sorts strings", func(t *testing.T) {
		input := []string{"c", "a", "b"}
		result := sortedCopy(input)

		expected := []string{"a", "b", "c"}
		if len(result) != len(expected) {
			t.Fatalf("Expected %d elements, got %d", len(expected), len(result))
		}

		for i := range expected {
			if result[i] != expected[i] {
				t.Errorf("At index %d: expected %s, got %s", i, expected[i], result[i])
			}
		}

		// Original should be unchanged
		if input[0] != "c" {
			t.Error("Original slice was modified")
		}
	})
}

func TestCacheHelperMethods(t *testing.T) {
	t.Run("SetArtifacts and GetComponentHash", func(t *testing.T) {
		c := New()
		c.Components["test.comp"] = &Component{
			Hash:      "hash123",
			Artifacts: []string{},
		}

		// Set artifacts
		c.SetArtifacts("test.comp", []string{"file1.ts", "file2.ts"})

		// Verify artifacts were set
		if comp, ok := c.Components["test.comp"]; ok {
			if len(comp.Artifacts) != 2 {
				t.Errorf("Expected 2 artifacts, got %d", len(comp.Artifacts))
			}
		} else {
			t.Error("Component not found")
		}

		// Get component hash
		hash := c.GetComponentHash("test.comp")
		if hash != "hash123" {
			t.Errorf("Expected hash hash123, got %s", hash)
		}

		// Non-existent component
		hash = c.GetComponentHash("nonexistent")
		if hash != "" {
			t.Errorf("Expected empty hash for nonexistent component, got %s", hash)
		}
	})

	t.Run("HasComponent", func(t *testing.T) {
		c := New()
		c.Components["test.comp"] = &Component{Hash: "hash123"}

		if !c.HasComponent("test.comp") {
			t.Error("Expected HasComponent to return true")
		}

		if c.HasComponent("nonexistent") {
			t.Error("Expected HasComponent to return false for nonexistent component")
		}
	})
}
