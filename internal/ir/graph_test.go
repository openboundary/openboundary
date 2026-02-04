package ir

import (
	"testing"

	"github.com/stack-bound/stack-bound/internal/parser"
)

func TestIR_DetectCycles_NoCycles(t *testing.T) {
	spec := &parser.Spec{
		Components: []parser.Component{
			{ID: "comp.a", Kind: "http.server", Spec: map[string]interface{}{"framework": "hono", "port": 3000}},
			{ID: "comp.b", Kind: "postgres", Spec: map[string]interface{}{"provider": "drizzle", "schema": "./s.ts"}},
		},
	}

	b := NewBuilder()
	ir, _ := b.Build(spec)

	cycles := ir.DetectCycles()
	if len(cycles) != 0 {
		t.Errorf("DetectCycles() found %d cycles, expected 0", len(cycles))
	}
}

func TestIR_DetectCycles_WithCycle(t *testing.T) {
	// Create a cycle: A -> B -> A
	ir := New(&parser.Spec{})

	compA := &Component{ID: "comp.a", Kind: KindHTTPServer, Dependencies: []*Component{}, Dependents: []*Component{}}
	compB := &Component{ID: "comp.b", Kind: KindPostgres, Dependencies: []*Component{}, Dependents: []*Component{}}

	// A depends on B, B depends on A
	compA.Dependencies = append(compA.Dependencies, compB)
	compB.Dependencies = append(compB.Dependencies, compA)

	ir.Components["comp.a"] = compA
	ir.Components["comp.b"] = compB

	cycles := ir.DetectCycles()
	if len(cycles) == 0 {
		t.Error("DetectCycles() did not detect cycle")
	}
}

func TestIR_TopologicalSort_Success(t *testing.T) {
	spec := &parser.Spec{
		Components: []parser.Component{
			{ID: "postgres.primary", Kind: "postgres", Spec: map[string]interface{}{"provider": "drizzle", "schema": "./s.ts"}},
			{
				ID:   "http.server.api",
				Kind: "http.server",
				Spec: map[string]interface{}{
					"framework":  "hono",
					"port":       3000,
					"depends_on": []interface{}{"postgres.primary"},
				},
			},
		},
	}

	b := NewBuilder()
	ir, _ := b.Build(spec)

	sorted, err := ir.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort() error = %v", err)
	}

	if len(sorted) != 2 {
		t.Errorf("TopologicalSort() returned %d components, expected 2", len(sorted))
	}

	// postgres should come before http.server since http.server depends on postgres
	postgresIdx := -1
	serverIdx := -1
	for i, c := range sorted {
		if c.ID == "postgres.primary" {
			postgresIdx = i
		}
		if c.ID == "http.server.api" {
			serverIdx = i
		}
	}

	if postgresIdx >= serverIdx {
		t.Errorf("postgres should come before http.server in topological order")
	}
}

func TestIR_TopologicalSort_WithCycle(t *testing.T) {
	ir := New(&parser.Spec{})

	compA := &Component{ID: "comp.a", Kind: KindHTTPServer, Dependencies: []*Component{}, Dependents: []*Component{}}
	compB := &Component{ID: "comp.b", Kind: KindPostgres, Dependencies: []*Component{}, Dependents: []*Component{}}

	compA.Dependencies = append(compA.Dependencies, compB)
	compB.Dependencies = append(compB.Dependencies, compA)

	ir.Components["comp.a"] = compA
	ir.Components["comp.b"] = compB

	_, err := ir.TopologicalSort()
	if err == nil {
		t.Error("TopologicalSort() expected error for cycle")
	}

	cycleErr, ok := err.(*CycleError)
	if !ok {
		t.Errorf("TopologicalSort() error type = %T, expected *CycleError", err)
	}

	if cycleErr.Error() == "" {
		t.Error("CycleError.Error() is empty")
	}
}

func TestIR_DependenciesOf(t *testing.T) {
	spec := &parser.Spec{
		Components: []parser.Component{
			{ID: "postgres.primary", Kind: "postgres", Spec: map[string]interface{}{"provider": "drizzle", "schema": "./s.ts"}},
			{
				ID:   "http.server.api",
				Kind: "http.server",
				Spec: map[string]interface{}{
					"framework":  "hono",
					"port":       3000,
					"depends_on": []interface{}{"postgres.primary"},
				},
			},
		},
	}

	b := NewBuilder()
	ir, _ := b.Build(spec)

	t.Run("has dependencies", func(t *testing.T) {
		deps, err := ir.DependenciesOf("http.server.api")
		if err != nil {
			t.Fatalf("DependenciesOf() error = %v", err)
		}
		if len(deps) != 1 {
			t.Errorf("DependenciesOf() returned %d dependencies, expected 1", len(deps))
		}
		if deps[0].ID != "postgres.primary" {
			t.Errorf("DependenciesOf()[0].ID = %q, expected %q", deps[0].ID, "postgres.primary")
		}
	})

	t.Run("no dependencies", func(t *testing.T) {
		deps, err := ir.DependenciesOf("postgres.primary")
		if err != nil {
			t.Fatalf("DependenciesOf() error = %v", err)
		}
		if len(deps) != 0 {
			t.Errorf("DependenciesOf() returned %d dependencies, expected 0", len(deps))
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := ir.DependenciesOf("nonexistent")
		if err == nil {
			t.Error("DependenciesOf() expected error for nonexistent component")
		}
		_, ok := err.(*ComponentNotFoundError)
		if !ok {
			t.Errorf("error type = %T, expected *ComponentNotFoundError", err)
		}
	})
}

func TestIR_DependentsOf(t *testing.T) {
	spec := &parser.Spec{
		Components: []parser.Component{
			{ID: "postgres.primary", Kind: "postgres", Spec: map[string]interface{}{"provider": "drizzle", "schema": "./s.ts"}},
			{
				ID:   "http.server.api",
				Kind: "http.server",
				Spec: map[string]interface{}{
					"framework":  "hono",
					"port":       3000,
					"depends_on": []interface{}{"postgres.primary"},
				},
			},
		},
	}

	b := NewBuilder()
	ir, _ := b.Build(spec)

	t.Run("has dependents", func(t *testing.T) {
		dependents, err := ir.DependentsOf("postgres.primary")
		if err != nil {
			t.Fatalf("DependentsOf() error = %v", err)
		}
		if len(dependents) != 1 {
			t.Errorf("DependentsOf() returned %d dependents, expected 1", len(dependents))
		}
	})

	t.Run("no dependents", func(t *testing.T) {
		dependents, err := ir.DependentsOf("http.server.api")
		if err != nil {
			t.Fatalf("DependentsOf() error = %v", err)
		}
		if len(dependents) != 0 {
			t.Errorf("DependentsOf() returned %d dependents, expected 0", len(dependents))
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := ir.DependentsOf("nonexistent")
		if err == nil {
			t.Error("DependentsOf() expected error for nonexistent component")
		}
	})
}

func TestCycleError_Error(t *testing.T) {
	tests := []struct {
		name     string
		cycles   [][]string
		contains string
	}{
		{
			name:     "empty cycles",
			cycles:   [][]string{},
			contains: "dependency cycle detected",
		},
		{
			name:     "with cycle",
			cycles:   [][]string{{"a", "b", "c"}},
			contains: "a -> b -> c -> a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &CycleError{Cycles: tt.cycles}
			msg := err.Error()
			if msg == "" {
				t.Error("Error() returned empty string")
			}
		})
	}
}

func TestComponentNotFoundError_Error(t *testing.T) {
	err := &ComponentNotFoundError{ID: "test.comp"}
	msg := err.Error()
	if msg != "component not found: test.comp" {
		t.Errorf("Error() = %q", msg)
	}
}

func TestExtractCycle(t *testing.T) {
	tests := []struct {
		name     string
		path     []string
		targetID string
		expected []string
	}{
		{
			name:     "cycle at start",
			path:     []string{"a", "b", "c"},
			targetID: "a",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "cycle in middle",
			path:     []string{"a", "b", "c"},
			targetID: "b",
			expected: []string{"b", "c"},
		},
		{
			name:     "target not in path",
			path:     []string{"a", "b", "c"},
			targetID: "d",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCycle(tt.path, tt.targetID)
			if len(got) != len(tt.expected) {
				t.Errorf("extractCycle() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestFormatCycle(t *testing.T) {
	tests := []struct {
		cycle    []string
		expected string
	}{
		{[]string{}, ""},
		{[]string{"a"}, "a -> a"},
		{[]string{"a", "b"}, "a -> b -> a"},
		{[]string{"a", "b", "c"}, "a -> b -> c -> a"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := formatCycle(tt.cycle)
			if got != tt.expected {
				t.Errorf("formatCycle(%v) = %q, expected %q", tt.cycle, got, tt.expected)
			}
		})
	}
}
