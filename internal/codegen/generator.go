// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package codegen provides code generation from the IR.
package codegen

import "github.com/openboundary/openboundary/internal/ir"

// Generator is the interface for code generators.
type Generator interface {
	// Name returns the generator name.
	Name() string

	// Generate produces code from the IR.
	Generate(i *ir.IR) (*Output, error)
}

// Output represents the generated code output.
type Output struct {
	// Files maps relative paths to file contents.
	Files map[string][]byte
}

// NewOutput creates a new Output.
func NewOutput() *Output {
	return &Output{
		Files: make(map[string][]byte),
	}
}

// AddFile adds a file to the output.
func (o *Output) AddFile(path string, content []byte) {
	o.Files[path] = content
}

