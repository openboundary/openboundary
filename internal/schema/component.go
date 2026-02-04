// Package schema defines component kinds and their validation schemas.
package schema

import "slices"

// Kind represents a component kind.
type Kind string

// Known component kinds.
const (
	KindHTTPServer Kind = "http.server"
	KindMiddleware Kind = "middleware"
	KindPostgres   Kind = "postgres"
	KindUsecase    Kind = "usecase"
)

// AllKinds returns all known component kinds.
func AllKinds() []Kind {
	return []Kind{
		KindHTTPServer,
		KindMiddleware,
		KindPostgres,
		KindUsecase,
	}
}

// IsValidKind checks if the given kind is known.
func IsValidKind(k Kind) bool {
	return slices.Contains(AllKinds(), k)
}

// Schema defines the validation interface for component specs.
type Schema interface {
	// Kind returns the component kind this schema validates.
	Kind() Kind

	// Validate validates the component spec.
	Validate(spec map[string]interface{}) error
}

// Registry holds all registered schemas.
type Registry struct {
	schemas map[Kind]Schema
}

// NewRegistry creates a new schema registry.
func NewRegistry() *Registry {
	return &Registry{
		schemas: make(map[Kind]Schema),
	}
}

// Register adds a schema to the registry.
func (r *Registry) Register(s Schema) {
	r.schemas[s.Kind()] = s
}

// Get returns the schema for the given kind.
func (r *Registry) Get(k Kind) (Schema, bool) {
	s, ok := r.schemas[k]
	return s, ok
}

// DefaultRegistry returns a registry with all default schemas.
func DefaultRegistry() *Registry {
	reg := NewRegistry()
	// TODO: Register all default schemas
	return reg
}
