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

// Registry holds all registered generators.
type Registry struct {
	generators map[string]Generator
}

// NewRegistry creates a new generator registry.
func NewRegistry() *Registry {
	return &Registry{
		generators: make(map[string]Generator),
	}
}

// Register adds a generator to the registry.
func (r *Registry) Register(g Generator) {
	r.generators[g.Name()] = g
}

// Get returns a generator by name.
func (r *Registry) Get(name string) (Generator, bool) {
	g, ok := r.generators[name]
	return g, ok
}

// All returns all registered generators.
func (r *Registry) All() []Generator {
	gens := make([]Generator, 0, len(r.generators))
	for _, g := range r.generators {
		gens = append(gens, g)
	}
	return gens
}
