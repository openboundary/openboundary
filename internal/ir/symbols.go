// Copyright 2026 OpenBoundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package ir

import "fmt"

// SymbolTable tracks all named symbols in the specification.
type SymbolTable struct {
	symbols map[string]*Symbol
}

// Symbol represents a named entity in the specification.
type Symbol struct {
	Name      string
	Kind      Kind
	Component *Component
}

// NewSymbolTable creates a new symbol table.
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		symbols: make(map[string]*Symbol),
	}
}

// Define adds a symbol to the table.
func (t *SymbolTable) Define(name string, kind Kind, comp *Component) error {
	if existing, ok := t.symbols[name]; ok {
		return fmt.Errorf("symbol %q already defined as %s", name, existing.Kind)
	}
	t.symbols[name] = &Symbol{
		Name:      name,
		Kind:      kind,
		Component: comp,
	}
	return nil
}

// Lookup returns a symbol by name.
func (t *SymbolTable) Lookup(name string) (*Symbol, bool) {
	sym, ok := t.symbols[name]
	return sym, ok
}

// All returns all symbols in the table.
func (t *SymbolTable) All() []*Symbol {
	result := make([]*Symbol, 0, len(t.symbols))
	for _, sym := range t.symbols {
		result = append(result, sym)
	}
	return result
}

// Len returns the number of symbols in the table.
func (t *SymbolTable) Len() int {
	return len(t.symbols)
}
