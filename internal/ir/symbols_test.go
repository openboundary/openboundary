// Copyright 2026 OpenBoundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package ir

import (
	"testing"
)

func TestNewSymbolTable(t *testing.T) {
	st := NewSymbolTable()
	if st == nil {
		t.Fatal("NewSymbolTable() returned nil")
	}
	if st.symbols == nil {
		t.Error("NewSymbolTable() symbols map is nil")
	}
	if st.Len() != 0 {
		t.Errorf("NewSymbolTable() Len() = %d, expected 0", st.Len())
	}
}

func TestSymbolTable_Define(t *testing.T) {
	st := NewSymbolTable()
	comp := &Component{ID: "test.comp", Kind: KindHTTPServer}

	err := st.Define("test.comp", KindHTTPServer, comp)
	if err != nil {
		t.Errorf("Define() error = %v", err)
	}

	if st.Len() != 1 {
		t.Errorf("Len() = %d, expected 1", st.Len())
	}
}

func TestSymbolTable_Define_Duplicate(t *testing.T) {
	st := NewSymbolTable()
	comp1 := &Component{ID: "test.comp", Kind: KindHTTPServer}
	comp2 := &Component{ID: "test.comp", Kind: KindPostgres}

	_ = st.Define("test.comp", KindHTTPServer, comp1)
	err := st.Define("test.comp", KindPostgres, comp2)

	if err == nil {
		t.Error("Define() expected error for duplicate, got nil")
	}
}

func TestSymbolTable_Lookup(t *testing.T) {
	st := NewSymbolTable()
	comp := &Component{ID: "test.comp", Kind: KindHTTPServer}
	_ = st.Define("test.comp", KindHTTPServer, comp)

	t.Run("found", func(t *testing.T) {
		sym, ok := st.Lookup("test.comp")
		if !ok {
			t.Error("Lookup() returned false for existing symbol")
		}
		if sym == nil {
			t.Fatal("Lookup() returned nil symbol")
		}
		if sym.Name != "test.comp" {
			t.Errorf("Symbol.Name = %q, expected %q", sym.Name, "test.comp")
		}
		if sym.Kind != KindHTTPServer {
			t.Errorf("Symbol.Kind = %v, expected %v", sym.Kind, KindHTTPServer)
		}
		if sym.Component != comp {
			t.Error("Symbol.Component does not match")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, ok := st.Lookup("nonexistent")
		if ok {
			t.Error("Lookup() returned true for nonexistent symbol")
		}
	})
}

func TestSymbolTable_All(t *testing.T) {
	st := NewSymbolTable()

	t.Run("empty", func(t *testing.T) {
		all := st.All()
		if len(all) != 0 {
			t.Errorf("All() on empty table returned %d symbols", len(all))
		}
	})

	t.Run("with symbols", func(t *testing.T) {
		comp1 := &Component{ID: "comp.a", Kind: KindHTTPServer}
		comp2 := &Component{ID: "comp.b", Kind: KindPostgres}
		_ = st.Define("comp.a", KindHTTPServer, comp1)
		_ = st.Define("comp.b", KindPostgres, comp2)

		all := st.All()
		if len(all) != 2 {
			t.Errorf("All() returned %d symbols, expected 2", len(all))
		}
	})
}

func TestSymbolTable_Len(t *testing.T) {
	st := NewSymbolTable()

	if st.Len() != 0 {
		t.Errorf("Len() on empty table = %d", st.Len())
	}

	_ = st.Define("a", KindHTTPServer, &Component{})
	if st.Len() != 1 {
		t.Errorf("Len() after 1 define = %d", st.Len())
	}

	_ = st.Define("b", KindPostgres, &Component{})
	if st.Len() != 2 {
		t.Errorf("Len() after 2 defines = %d", st.Len())
	}
}
