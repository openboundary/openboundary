// Copyright 2026 OpenBoundary Contributors
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

// WriteStrategy controls how a file is written to disk.
type WriteStrategy int

const (
	// WriteAlways overwrites the file on every compile (default).
	WriteAlways WriteStrategy = iota
	// WriteOnce writes the file only if it does not already exist.
	WriteOnce
)

// OutputFile represents a single generated file with optional component association.
type OutputFile struct {
	Content     []byte
	ComponentID string        // Optional: which component this file belongs to (empty for shared files)
	Strategy    WriteStrategy // How the write stage should handle this file
}

// Output represents the generated code output.
type Output struct {
	// Files maps relative paths to file info.
	Files map[string]OutputFile
}

// NewOutput creates a new Output.
func NewOutput() *Output {
	return &Output{
		Files: make(map[string]OutputFile),
	}
}

// AddFile adds a file to the output without component association (shared file).
func (o *Output) AddFile(path string, content []byte) {
	o.Files[path] = OutputFile{
		Content:     content,
		ComponentID: "",
	}
}

// AddComponentFile adds a file to the output with component association.
func (o *Output) AddComponentFile(path string, content []byte, componentID string) {
	o.Files[path] = OutputFile{
		Content:     content,
		ComponentID: componentID,
	}
}

// AddWriteOnceFile adds a file that will only be written if it does not already exist on disk.
func (o *Output) AddWriteOnceFile(path string, content []byte) {
	o.Files[path] = OutputFile{
		Content:  content,
		Strategy: WriteOnce,
	}
}

// AddWriteOnceComponentFile adds a component file that will only be written if it does not already exist on disk.
func (o *Output) AddWriteOnceComponentFile(path string, content []byte, componentID string) {
	o.Files[path] = OutputFile{
		Content:     content,
		ComponentID: componentID,
		Strategy:    WriteOnce,
	}
}

