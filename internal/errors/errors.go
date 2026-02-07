// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package errors provides enhanced error handling with actionable guidance.
package errors

import (
	"fmt"
	"strings"
)

// UserError represents an error with actionable guidance for users.
type UserError struct {
	Title       string   // Clear, concise error title
	Context     string   // Why this error matters
	Solutions   []string // Ordered list of things to try
	DocsTopic   string   // Related docs topic (optional)
	Underlying  error    // Original error (optional)
}

// Error implements the error interface.
func (e *UserError) Error() string {
	var b strings.Builder

	b.WriteString("Error: ")
	b.WriteString(e.Title)
	b.WriteString("\n")

	if e.Context != "" {
		b.WriteString("\n")
		b.WriteString(e.Context)
		b.WriteString("\n")
	}

	if len(e.Solutions) > 0 {
		b.WriteString("\nTry these solutions:\n")
		for i, solution := range e.Solutions {
			fmt.Fprintf(&b, "%d. %s\n", i+1, solution)
		}
	}

	if e.DocsTopic != "" {
		fmt.Fprintf(&b, "\nFor more help: bound docs %s\n", e.DocsTopic)
	}

	if e.Underlying != nil {
		fmt.Fprintf(&b, "\nDetails: %v\n", e.Underlying)
	}

	return b.String()
}

// Unwrap returns the underlying error for error chain inspection.
func (e *UserError) Unwrap() error {
	return e.Underlying
}

// FileNotFoundError creates an error for missing spec files.
func FileNotFoundError(path string, err error) *UserError {
	return &UserError{
		Title:   fmt.Sprintf("Specification file not found: %s", path),
		Context: "The compiler needs a valid specification file to generate code.",
		Solutions: []string{
			"Check that the file path is correct",
			"Verify the file exists in the current directory",
			"Create a new project with: bound init <project-name>",
		},
		DocsTopic:  "init",
		Underlying: err,
	}
}

// InvalidYAMLError creates an error for YAML parsing failures.
func InvalidYAMLError(file string, err error) *UserError {
	return &UserError{
		Title:   "Failed to parse specification file",
		Context: "The YAML syntax in your spec file is invalid.",
		Solutions: []string{
			"Check for proper YAML indentation (use spaces, not tabs)",
			"Verify all strings with special characters are quoted",
			"Look for missing colons after keys",
			"Validate YAML syntax with: yamllint " + file,
		},
		DocsTopic:  "spec",
		Underlying: err,
	}
}

// SchemaValidationError creates an error for schema validation failures.
func SchemaValidationError(errors []string) *UserError {
	details := strings.Join(errors, "\n  - ")
	return &UserError{
		Title:   "Specification does not match the OpenBoundary schema",
		Context: "Your spec file has structural or type errors that prevent compilation.",
		Solutions: []string{
			"Review the validation errors below",
			"Check the OpenBoundary schema reference",
			"Ensure all required fields are present",
			"Verify component kinds are valid (http.server, middleware, postgres, usecase)",
		},
		DocsTopic: "spec",
		Underlying: fmt.Errorf("validation errors:\n  - %s", details),
	}
}

// GeneratorError creates an error for code generation failures.
func GeneratorError(generator string, err error) *UserError {
	return &UserError{
		Title:   fmt.Sprintf("Code generation failed: %s", generator),
		Context: "The generator encountered an error while creating artifacts.",
		Solutions: []string{
			"Check your spec file for missing or invalid component configuration",
			"Verify all referenced files (OpenAPI specs) exist and are valid",
			"Try with --force-regenerate to clear any cached state",
			"Report this issue if the error persists",
		},
		DocsTopic:  "generators",
		Underlying: err,
	}
}

// WriteError creates an error for file writing failures.
func WriteError(path string, err error) *UserError {
	return &UserError{
		Title:   fmt.Sprintf("Failed to write file: %s", path),
		Context: "The compiler could not write the generated file to disk.",
		Solutions: []string{
			"Check that you have write permissions in the output directory",
			"Verify there is enough disk space available",
			"Ensure the parent directory exists and is writable",
			"Check if the file is locked by another process",
		},
		Underlying: err,
	}
}

// PathTraversalError creates an error for path traversal attacks.
func PathTraversalError(path string) *UserError {
	return &UserError{
		Title:   "Path traversal detected",
		Context: fmt.Sprintf("The artifact path '%s' contains suspicious characters or patterns that could lead to path traversal attacks.", path),
		Solutions: []string{
			"Ensure artifact paths are relative and don't contain '..' or absolute paths",
			"Remove Unicode control characters from file names",
			"Check generator code for proper path sanitization",
		},
		DocsTopic: "security",
		Underlying: fmt.Errorf("unsafe path: %s", path),
	}
}

// CacheError wraps cache-related errors with context.
func CacheError(operation string, err error) *UserError {
	return &UserError{
		Title:   fmt.Sprintf("Cache %s failed", operation),
		Context: "The compiler encountered an error while working with the compilation cache.",
		Solutions: []string{
			"Try deleting the cache: rm -rf <output-dir>/.bound/cache.json",
			"Run with --no-cache to bypass caching",
			"Check file permissions in the output directory",
		},
		Underlying: err,
	}
}
