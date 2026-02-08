// Copyright 2026 OpenBoundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package ir provides the typed intermediate representation for code generation.
package ir

import (
	"fmt"
	"slices"

	"github.com/openboundary/openboundary/internal/openapi"
	"github.com/openboundary/openboundary/internal/parser"
)

// IR is the typed intermediate representation used for code generation.
// It contains resolved references and a complete dependency graph.
type IR struct {
	Spec       *parser.Spec
	Components map[string]*Component
	Edges      []Edge
	Symbols    *SymbolTable
	BaseDir    string // Base directory for resolving relative paths
}

// New creates a new IR from a parsed spec.
func New(spec *parser.Spec) *IR {
	return &IR{
		Spec:       spec,
		Components: make(map[string]*Component),
		Edges:      []Edge{},
		Symbols:    NewSymbolTable(),
	}
}

// Component represents a resolved component in the IR.
type Component struct {
	ID           string
	Kind         Kind
	Position     parser.Position
	Dependencies []*Component
	Dependents   []*Component

	// Kind-specific typed specs
	HTTPServer *HTTPServerSpec
	Middleware *MiddlewareSpec
	Postgres   *PostgresSpec
	Usecase    *UsecaseSpec
}

// Kind represents a component kind.
type Kind string

// Known component kinds.
// TODO: Make kinds extendable via a KindPlugin interface so each kind ships its
// own spec parser, reference resolver, validator, and schema fragment. Holding
// off until a 5th kind (or a 3rd-party kind) forces the design — the abstraction
// boundary between kinds isn't clear enough yet with only four first-party kinds.
const (
	KindHTTPServer Kind = "http.server"
	KindMiddleware Kind = "middleware"
	KindPostgres   Kind = "postgres"
	KindUsecase    Kind = "usecase"
)

// ParseKind converts a string to a Kind.
func ParseKind(s string) (Kind, error) {
	switch s {
	case string(KindHTTPServer):
		return KindHTTPServer, nil
	case string(KindMiddleware):
		return KindMiddleware, nil
	case string(KindPostgres):
		return KindPostgres, nil
	case string(KindUsecase):
		return KindUsecase, nil
	default:
		return "", fmt.Errorf("unknown kind: %s", s)
	}
}

// AllKinds returns all known component kinds.
func AllKinds() []Kind {
	return []Kind{KindHTTPServer, KindMiddleware, KindPostgres, KindUsecase}
}

// IsValidKind checks if the given kind is known.
func IsValidKind(k Kind) bool {
	return slices.Contains(AllKinds(), k)
}

// HTTPServerSpec contains typed fields for http.server components.
type HTTPServerSpec struct {
	Framework  string
	Port       int
	OpenAPI    string
	Middleware []string
	DependsOn  []string

	// ParsedOpenAPI contains the parsed OpenAPI document (populated during build phase).
	ParsedOpenAPI *openapi.Document
}

// MiddlewareSpec contains typed fields for middleware components.
type MiddlewareSpec struct {
	Provider  string // todo - leaky abstraction - consider subtypes for authn & authz
	Config    string
	Model     string
	Policy    string
	DependsOn []string
}

// PostgresSpec contains typed fields for postgres components.
type PostgresSpec struct {
	Provider string
	Schema   string
}

// UsecaseSpec contains typed fields for usecase components.
type UsecaseSpec struct {
	BindsTo            string
	Middleware         []string
	Goal               string
	Actor              string
	Preconditions      []string
	AcceptanceCriteria []string
	Postconditions     []string

	// Binding contains the parsed binding information (populated during build phase).
	Binding *Binding
}

// Binding represents a parsed binds_to value with resolved references.
type Binding struct {
	ServerID  string             // The server component ID
	Method    string             // HTTP method (GET, POST, etc.)
	Path      string             // URL path (e.g., /users/{id})
	Operation *openapi.Operation // The resolved OpenAPI operation (may be nil if not found)
}

// Edge represents a dependency edge between components.
type Edge struct {
	From *Component
	To   *Component
	Type EdgeType
}

// EdgeType represents the type of dependency.
type EdgeType string

// Known edge types.
const (
	EdgeTypeRef        EdgeType = "ref"
	EdgeTypeDependency EdgeType = "dependency"
	EdgeTypeMiddleware EdgeType = "middleware"
	EdgeTypeBinding    EdgeType = "binding"
)
