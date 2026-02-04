package ir

import (
	"fmt"

	"github.com/stack-bound/stack-bound/internal/openapi"
	"github.com/stack-bound/stack-bound/internal/parser"
)

// Builder builds a typed IR from a parsed spec.
type Builder struct {
	baseDir string // Base directory for resolving relative paths
}

// NewBuilder creates a new IR builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// WithBaseDir sets the base directory for resolving relative paths (e.g., OpenAPI files).
func (b *Builder) WithBaseDir(dir string) *Builder {
	b.baseDir = dir
	return b
}

// Build creates a typed IR from the given spec.
// It resolves all references and builds the dependency graph.
func (b *Builder) Build(spec *parser.Spec) (*IR, []error) {
	ir := New(spec)
	ir.BaseDir = b.baseDir
	var errs []error

	// Phase 1: Create components and populate symbol table
	for i := range spec.Components {
		comp := &spec.Components[i]
		kind, err := ParseKind(comp.Kind)
		if err != nil {
			errs = append(errs, fmt.Errorf("component %q: %w", comp.ID, err))
			continue
		}

		irComp := &Component{
			ID:           comp.ID,
			Kind:         kind,
			Position:     comp.Pos(),
			Dependencies: []*Component{},
			Dependents:   []*Component{},
		}

		// Parse kind-specific spec
		b.parseComponentSpec(irComp, comp.Spec)

		ir.Components[comp.ID] = irComp

		if err := ir.Symbols.Define(comp.ID, kind, irComp); err != nil {
			errs = append(errs, err)
		}
	}

	// If we had errors creating components, don't try to resolve references
	if len(errs) > 0 {
		return ir, errs
	}

	// Phase 2: Parse OpenAPI specs for http.server components
	openAPIErrs := b.parseOpenAPISpecs(ir)
	errs = append(errs, openAPIErrs...)

	// Phase 3: Resolve references and build edges
	for _, comp := range ir.Components {
		refErrs := b.resolveReferences(ir, comp)
		errs = append(errs, refErrs...)
	}

	// Phase 4: Link usecases to OpenAPI operations
	linkErrs := b.linkUsecasesToOperations(ir)
	errs = append(errs, linkErrs...)

	return ir, errs
}

// parseOpenAPISpecs parses OpenAPI specs for all http.server components.
func (b *Builder) parseOpenAPISpecs(ir *IR) []error {
	var errs []error
	oaParser := openapi.NewParser(b.baseDir)

	for _, comp := range ir.Components {
		if comp.Kind != KindHTTPServer || comp.HTTPServer == nil {
			continue
		}

		if comp.HTTPServer.OpenAPI == "" {
			continue
		}

		doc, err := oaParser.ParseFile(comp.HTTPServer.OpenAPI)
		if err != nil {
			errs = append(errs, fmt.Errorf("component %q: failed to parse OpenAPI spec %q: %w",
				comp.ID, comp.HTTPServer.OpenAPI, err))
			continue
		}

		comp.HTTPServer.ParsedOpenAPI = doc
	}

	return errs
}

// linkUsecasesToOperations parses binds_to and links usecases to their OpenAPI operations.
func (b *Builder) linkUsecasesToOperations(ir *IR) []error {
	var errs []error

	for _, comp := range ir.Components {
		if comp.Kind != KindUsecase || comp.Usecase == nil {
			continue
		}

		if comp.Usecase.BindsTo == "" {
			continue
		}

		// Parse the binding
		serverID, method, path, err := openapi.ParseBinding(comp.Usecase.BindsTo)
		if err != nil {
			errs = append(errs, fmt.Errorf("component %q: invalid binds_to: %w", comp.ID, err))
			continue
		}

		binding := &Binding{
			ServerID: serverID,
			Method:   method,
			Path:     path,
		}

		// Look up the server component
		serverSym, ok := ir.Symbols.Lookup(serverID)
		if !ok {
			errs = append(errs, fmt.Errorf("component %q: server %q not found", comp.ID, serverID))
			continue
		}

		if serverSym.Kind != KindHTTPServer {
			errs = append(errs, fmt.Errorf("component %q: %q is not an http.server", comp.ID, serverID))
			continue
		}

		serverComp := serverSym.Component
		if serverComp.HTTPServer == nil || serverComp.HTTPServer.ParsedOpenAPI == nil {
			// Server has no OpenAPI spec, binding is still valid but no operation resolution
			comp.Usecase.Binding = binding
			continue
		}

		// Look up the operation in the server's OpenAPI spec
		opKey := openapi.OperationKey(method, path)
		op, ok := serverComp.HTTPServer.ParsedOpenAPI.Operations[opKey]
		if !ok {
			errs = append(errs, fmt.Errorf("component %q: operation %s not found in %q's OpenAPI spec",
				comp.ID, opKey, serverID))
			continue
		}

		binding.Operation = op
		comp.Usecase.Binding = binding
	}

	return errs
}

// parseComponentSpec parses the untyped spec into typed fields.
// Note: Unknown kinds are filtered out before this function is called,
// so the switch is exhaustive for all valid kinds.
func (b *Builder) parseComponentSpec(comp *Component, spec map[string]interface{}) {
	switch comp.Kind {
	case KindHTTPServer:
		b.parseHTTPServerSpec(comp, spec)
	case KindMiddleware:
		b.parseMiddlewareSpec(comp, spec)
	case KindPostgres:
		b.parsePostgresSpec(comp, spec)
	case KindUsecase:
		b.parseUsecaseSpec(comp, spec)
	}
}

func (b *Builder) parseHTTPServerSpec(comp *Component, spec map[string]interface{}) {
	s := &HTTPServerSpec{}

	if v, ok := spec["framework"].(string); ok {
		s.Framework = v
	}
	if v, ok := spec["port"].(int); ok {
		s.Port = v
	} else if v, ok := spec["port"].(float64); ok {
		s.Port = int(v)
	}
	if v, ok := spec["openapi"].(string); ok {
		s.OpenAPI = v
	}
	if v, ok := spec["middleware"].([]interface{}); ok {
		s.Middleware = toStringSlice(v)
	}
	if v, ok := spec["depends_on"].([]interface{}); ok {
		s.DependsOn = toStringSlice(v)
	}

	comp.HTTPServer = s
}

func (b *Builder) parseMiddlewareSpec(comp *Component, spec map[string]interface{}) {
	s := &MiddlewareSpec{}

	if v, ok := spec["provider"].(string); ok {
		s.Provider = v
	}
	if v, ok := spec["config"].(string); ok {
		s.Config = v
	}
	if v, ok := spec["model"].(string); ok {
		s.Model = v
	}
	if v, ok := spec["policy"].(string); ok {
		s.Policy = v
	}
	if v, ok := spec["depends_on"].([]interface{}); ok {
		s.DependsOn = toStringSlice(v)
	}

	comp.Middleware = s
}

func (b *Builder) parsePostgresSpec(comp *Component, spec map[string]interface{}) {
	s := &PostgresSpec{}

	if v, ok := spec["provider"].(string); ok {
		s.Provider = v
	}
	if v, ok := spec["schema"].(string); ok {
		s.Schema = v
	}

	comp.Postgres = s
}

func (b *Builder) parseUsecaseSpec(comp *Component, spec map[string]interface{}) {
	s := &UsecaseSpec{}

	if v, ok := spec["binds_to"].(string); ok {
		s.BindsTo = v
	}
	if v, ok := spec["middleware"].([]interface{}); ok {
		s.Middleware = toStringSlice(v)
	}
	if v, ok := spec["goal"].(string); ok {
		s.Goal = v
	}
	if v, ok := spec["actor"].(string); ok {
		s.Actor = v
	}
	if v, ok := spec["preconditions"].([]interface{}); ok {
		s.Preconditions = toStringSlice(v)
	}
	if v, ok := spec["acceptance_criteria"].([]interface{}); ok {
		s.AcceptanceCriteria = toStringSlice(v)
	}
	if v, ok := spec["postconditions"].([]interface{}); ok {
		s.Postconditions = toStringSlice(v)
	}

	comp.Usecase = s
}

// resolveReferences resolves all references from a component and creates edges.
func (b *Builder) resolveReferences(ir *IR, comp *Component) []error {
	var errs []error

	switch comp.Kind {
	case KindHTTPServer:
		if comp.HTTPServer != nil {
			for _, ref := range comp.HTTPServer.Middleware {
				if err := b.addEdge(ir, comp, ref, EdgeTypeMiddleware); err != nil {
					errs = append(errs, err)
				}
			}
			for _, ref := range comp.HTTPServer.DependsOn {
				if err := b.addEdge(ir, comp, ref, EdgeTypeDependency); err != nil {
					errs = append(errs, err)
				}
			}
		}
	case KindMiddleware:
		if comp.Middleware != nil {
			for _, ref := range comp.Middleware.DependsOn {
				if err := b.addEdge(ir, comp, ref, EdgeTypeDependency); err != nil {
					errs = append(errs, err)
				}
			}
		}
	case KindUsecase:
		if comp.Usecase != nil {
			// Parse binds_to to extract server reference
			if comp.Usecase.BindsTo != "" {
				serverID := extractServerFromBinding(comp.Usecase.BindsTo)
				if serverID != "" {
					if err := b.addEdge(ir, comp, serverID, EdgeTypeBinding); err != nil {
						errs = append(errs, err)
					}
				}
			}
			for _, ref := range comp.Usecase.Middleware {
				if err := b.addEdge(ir, comp, ref, EdgeTypeMiddleware); err != nil {
					errs = append(errs, err)
				}
			}
		}
	}

	return errs
}

func (b *Builder) addEdge(ir *IR, from *Component, toRef string, edgeType EdgeType) error {
	sym, ok := ir.Symbols.Lookup(toRef)
	if !ok {
		return fmt.Errorf("unresolved reference %q in component %q", toRef, from.ID)
	}

	to := sym.Component
	from.Dependencies = append(from.Dependencies, to)
	to.Dependents = append(to.Dependents, from)

	ir.Edges = append(ir.Edges, Edge{
		From: from,
		To:   to,
		Type: edgeType,
	})

	return nil
}

// extractServerFromBinding extracts the server ID from a binds_to value.
// Format: server-id:METHOD:/path
func extractServerFromBinding(bindsTo string) string {
	for i, c := range bindsTo {
		if c == ':' {
			return bindsTo[:i]
		}
	}
	return ""
}

// toStringSlice converts an interface slice to a string slice.
// Non-string items are silently skipped. This is intentional to allow
// YAML parsing flexibility, but callers should be aware that invalid
// entries will not produce errors here - validate at the schema level.
func toStringSlice(v []interface{}) []string {
	result := make([]string, 0, len(v))
	for _, item := range v {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
