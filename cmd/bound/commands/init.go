// Copyright 2026 OpenBoundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package commands

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/openboundary/openboundary/templates"
)

func Init(projectName, template string) error {
	// Reject path traversal or absolute paths in project name.
	if filepath.IsAbs(projectName) || strings.Contains(projectName, "..") || strings.ContainsAny(projectName, `/\`) {
		return fmt.Errorf("invalid project name %q: must be a simple directory name", projectName)
	}

	// Verify the template exists in the embedded filesystem.
	entries, err := fs.ReadDir(templates.FS, template)
	if err != nil {
		return fmt.Errorf("unknown template %q: available templates are blank, basic", template)
	}

	if len(entries) == 0 {
		return fmt.Errorf("template %q is empty", template)
	}

	// Ensure project directory does not already exist.
	if _, err := os.Stat(projectName); err == nil {
		return fmt.Errorf("directory %q already exists", projectName)
	}

	if err := os.MkdirAll(projectName, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	count := 0
	err = fs.WalkDir(templates.FS, template, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Get path relative to the template root.
		relPath, _ := filepath.Rel(template, path)
		if relPath == "." {
			return nil
		}

		destPath := filepath.Join(projectName, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		content, err := fs.ReadFile(templates.FS, path)
		if err != nil {
			return fmt.Errorf("failed to read template file %s: %w", path, err)
		}

		// Replace placeholder project name in spec.yaml.
		if filepath.Base(path) == "spec.yaml" {
			content = []byte(strings.ReplaceAll(string(content), "name: user-api", "name: "+projectName))
			content = []byte(strings.ReplaceAll(string(content), "name: Blank Project", "name: "+projectName))
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", relPath, err)
		}

		fmt.Printf("  → %s\n", relPath)
		count++
		return nil
	})

	if err != nil {
		return err
	}

	fmt.Printf("\n✓ Initialized %s project with %d files\n", template, count)
	return nil
}
