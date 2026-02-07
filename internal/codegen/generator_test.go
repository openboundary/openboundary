// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package codegen

import (
	"testing"

	"github.com/openboundary/openboundary/internal/ir"
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
