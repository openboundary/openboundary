package codegen

import (
	"testing"

	"github.com/stack-bound/stack-bound/internal/ir"
)

func TestNewOutput(t *testing.T) {
	o := NewOutput()
	if o == nil {
		t.Fatal("NewOutput() returned nil")
	}
	if o.Files == nil {
		t.Error("NewOutput().Files is nil")
	}
}

func TestOutput_AddFile(t *testing.T) {
	o := NewOutput()
	content := []byte("test content")
	path := "test/file.ts"

	o.AddFile(path, content)

	if got, ok := o.Files[path]; !ok {
		t.Error("AddFile() did not add file")
	} else if string(got) != string(content) {
		t.Errorf("AddFile() content = %q, expected %q", string(got), string(content))
	}
}

func TestOutput_AddFile_Overwrite(t *testing.T) {
	o := NewOutput()
	path := "test/file.ts"

	o.AddFile(path, []byte("first"))
	o.AddFile(path, []byte("second"))

	if string(o.Files[path]) != "second" {
		t.Errorf("AddFile() did not overwrite, got %q", string(o.Files[path]))
	}
}

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry() returned nil")
	}
	if r.generators == nil {
		t.Error("NewRegistry().generators is nil")
	}
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	mock := &mockGenerator{name: "test-gen"}

	r.Register(mock)

	got, ok := r.Get("test-gen")
	if !ok {
		t.Error("Get() returned false for registered generator")
	}
	if got != mock {
		t.Error("Get() returned wrong generator")
	}
}

func TestRegistry_Get_NotFound(t *testing.T) {
	r := NewRegistry()

	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("Get() returned true for unregistered generator")
	}
}

func TestRegistry_All(t *testing.T) {
	r := NewRegistry()
	mock1 := &mockGenerator{name: "gen1"}
	mock2 := &mockGenerator{name: "gen2"}

	r.Register(mock1)
	r.Register(mock2)

	all := r.All()
	if len(all) != 2 {
		t.Errorf("All() returned %d generators, expected 2", len(all))
	}
}

func TestRegistry_All_Empty(t *testing.T) {
	r := NewRegistry()
	all := r.All()
	if len(all) != 0 {
		t.Errorf("All() returned %d generators, expected 0", len(all))
	}
}

// mockGenerator implements Generator for testing
type mockGenerator struct {
	name   string
	output *Output
	err    error
}

func (m *mockGenerator) Name() string {
	return m.name
}

func (m *mockGenerator) Generate(i *ir.IR) (*Output, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.output != nil {
		return m.output, nil
	}
	return NewOutput(), nil
}
