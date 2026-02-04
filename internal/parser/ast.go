// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package parser provides YAML parsing with position tracking and AST definitions.
package parser

// Position tracks the location of a node in the source file.
type Position struct {
	File   string // Source file path
	Line   int    // 1-indexed line number
	Column int    // 1-indexed column number
}

// Node is the base interface for all AST nodes.
type Node interface {
	Pos() Position
}

// Spec represents the top-level specification.
type Spec struct {
	Version     string      `yaml:"version" json:"version"`
	Name        string      `yaml:"name" json:"name"`
	Description string      `yaml:"description,omitempty" json:"description,omitempty"`
	Components  []Component `yaml:"components" json:"components"`

	position Position
}

// Pos returns the position of the Spec in the source file.
func (s *Spec) Pos() Position {
	return s.position
}

// Component represents a single component in the specification.
// ID follows the pattern: type.subtype.name (e.g., "http.server.api", "middleware.authn")
type Component struct {
	ID   string                 `yaml:"id" json:"id"`
	Kind string                 `yaml:"kind" json:"kind"`
	Spec map[string]interface{} `yaml:"spec" json:"spec"`

	position Position
}

// Pos returns the position of the Component in the source file.
func (c *Component) Pos() Position {
	return c.position
}

// WithPosition creates a new Position for the given file and location.
func WithPosition(file string, line, column int) Position {
	return Position{
		File:   file,
		Line:   line,
		Column: column,
	}
}
