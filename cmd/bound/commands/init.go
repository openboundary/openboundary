package commands

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/openboundary/openboundary/templates"
)

func Init(template string) error {
	templateDir := template

	// Verify the template exists in the embedded filesystem.
	entries, err := fs.ReadDir(templates.FS, templateDir)
	if err != nil {
		return fmt.Errorf("unknown template %q: available templates are blank, basic", template)
	}

	if len(entries) == 0 {
		return fmt.Errorf("template %q is empty", template)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	projectName := filepath.Base(cwd)

	count := 0
	err = fs.WalkDir(templates.FS, templateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Get path relative to the template root.
		relPath, _ := filepath.Rel(templateDir, path)
		if relPath == "." {
			return nil
		}

		destPath := filepath.Join(cwd, relPath)

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
