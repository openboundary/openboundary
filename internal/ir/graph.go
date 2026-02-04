package ir

// DetectCycles returns any cycles found in the dependency graph.
func (ir *IR) DetectCycles() [][]string {
	var cycles [][]string
	visited := make(map[string]bool)
	inStack := make(map[string]bool)
	path := make([]string, 0)

	var dfs func(comp *Component) bool
	dfs = func(comp *Component) bool {
		visited[comp.ID] = true
		inStack[comp.ID] = true
		path = append(path, comp.ID)

		for _, dep := range comp.Dependencies {
			if !visited[dep.ID] {
				if dfs(dep) {
					return true
				}
			} else if inStack[dep.ID] {
				// Found a cycle - extract it from path
				cycle := extractCycle(path, dep.ID)
				cycles = append(cycles, cycle)
				return true
			}
		}

		path = path[:len(path)-1]
		inStack[comp.ID] = false
		return false
	}

	for _, comp := range ir.Components {
		if !visited[comp.ID] {
			dfs(comp)
		}
	}

	return cycles
}

// extractCycle extracts the cycle from the path starting at targetID.
func extractCycle(path []string, targetID string) []string {
	for i, id := range path {
		if id == targetID {
			cycle := make([]string, len(path)-i)
			copy(cycle, path[i:])
			return cycle
		}
	}
	return nil
}

// TopologicalSort returns components in dependency order.
// Components with no dependencies come first.
func (ir *IR) TopologicalSort() ([]*Component, error) {
	cycles := ir.DetectCycles()
	if len(cycles) > 0 {
		return nil, &CycleError{Cycles: cycles}
	}

	visited := make(map[string]bool)
	result := make([]*Component, 0, len(ir.Components))

	var visit func(comp *Component)
	visit = func(comp *Component) {
		if visited[comp.ID] {
			return
		}
		visited[comp.ID] = true

		// Visit dependencies first
		for _, dep := range comp.Dependencies {
			visit(dep)
		}

		result = append(result, comp)
	}

	for _, comp := range ir.Components {
		visit(comp)
	}

	return result, nil
}

// CycleError indicates a dependency cycle was detected.
type CycleError struct {
	Cycles [][]string
}

func (e *CycleError) Error() string {
	if len(e.Cycles) == 0 {
		return "dependency cycle detected"
	}
	return "dependency cycle detected: " + formatCycle(e.Cycles[0])
}

func formatCycle(cycle []string) string {
	if len(cycle) == 0 {
		return ""
	}
	result := cycle[0]
	for i := 1; i < len(cycle); i++ {
		result += " -> " + cycle[i]
	}
	result += " -> " + cycle[0]
	return result
}

// DependenciesOf returns the direct dependencies of a component.
func (ir *IR) DependenciesOf(id string) ([]*Component, error) {
	comp, ok := ir.Components[id]
	if !ok {
		return nil, &ComponentNotFoundError{ID: id}
	}
	return comp.Dependencies, nil
}

// DependentsOf returns components that depend on the given component.
func (ir *IR) DependentsOf(id string) ([]*Component, error) {
	comp, ok := ir.Components[id]
	if !ok {
		return nil, &ComponentNotFoundError{ID: id}
	}
	return comp.Dependents, nil
}

// ComponentNotFoundError indicates a component was not found.
type ComponentNotFoundError struct {
	ID string
}

func (e *ComponentNotFoundError) Error() string {
	return "component not found: " + e.ID
}
