// Copyright 2026 OpenBoundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/openboundary/openboundary/internal/codegen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubStage struct {
	name string
	err  error
	ran  bool
}

func (s *stubStage) Name() string { return s.name }
func (s *stubStage) Run(_ *Context) error {
	s.ran = true
	return s.err
}

func TestPipeline_RunsAllStages(t *testing.T) {
	s1 := &stubStage{name: "first"}
	s2 := &stubStage{name: "second"}
	s3 := &stubStage{name: "third"}

	p := New(s1, s2, s3)
	err := p.Run(&Context{})

	require.NoError(t, err)
	assert.True(t, s1.ran)
	assert.True(t, s2.ran)
	assert.True(t, s3.ran)
}

func TestPipeline_StopsOnFirstError(t *testing.T) {
	s1 := &stubStage{name: "first"}
	s2 := &stubStage{name: "second", err: errors.New("stage 2 failed")}
	s3 := &stubStage{name: "third"}

	p := New(s1, s2, s3)
	err := p.Run(&Context{})

	require.Error(t, err)
	assert.Equal(t, "stage 2 failed", err.Error())
	assert.True(t, s1.ran)
	assert.True(t, s2.ran)
	assert.False(t, s3.ran, "third stage should not run after error")
}

func TestPipeline_EmptyPipeline(t *testing.T) {
	p := New()
	err := p.Run(&Context{})
	require.NoError(t, err)
}

func TestParseStage_InvalidFile(t *testing.T) {
	stage := Parse()
	assert.Equal(t, "parse", stage.Name())

	ctx := &Context{SpecPath: "/nonexistent/file.yaml"}
	err := stage.Run(ctx)
	require.Error(t, err)
	// Parse stage wraps the error with "parse error"
	assert.Contains(t, err.Error(), "parse error")
}

func TestParseStage_ValidFile(t *testing.T) {
	stage := Parse()
	ctx := &Context{SpecPath: "../../examples/basic/spec.yaml"}
	err := stage.Run(ctx)
	require.NoError(t, err)
	assert.NotNil(t, ctx.AST)
	assert.Equal(t, "user-api", ctx.AST.Name)
}

func TestValidateSchemaStage_Name(t *testing.T) {
	stage := ValidateSchema()
	assert.Equal(t, "validate-schema", stage.Name())
}

func TestBuildIRStage_Name(t *testing.T) {
	stage := BuildIR()
	assert.Equal(t, "build-ir", stage.Name())
}

func TestValidateIRStage_Name(t *testing.T) {
	stage := ValidateIR()
	assert.Equal(t, "validate-ir", stage.Name())
}

func TestGenerateStage_Name(t *testing.T) {
	stage := Generate(nil)
	assert.Equal(t, "generate", stage.Name())
}

func TestWriteStage_Name(t *testing.T) {
	stage := Write()
	assert.Equal(t, "write", stage.Name())
}

func TestWriteStage_PathTraversal(t *testing.T) {
	outDir := t.TempDir()

	tests := []struct {
		name string
		path string
	}{
		{"dot-dot prefix", "../etc/passwd"},
		{"dot-dot nested", "subdir/../../etc/passwd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stage := Write()
			ctx := &Context{
				OutputDir: outDir,
				Artifacts: []codegen.Artifact{
					{Path: tt.path, Content: []byte("malicious")},
				},
			}
			err := stage.Run(ctx)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "escapes output directory")
		})
	}
}

func TestWriteStage_ValidPaths(t *testing.T) {
	outDir := t.TempDir()

	stage := Write()
	ctx := &Context{
		OutputDir: outDir,
		Artifacts: []codegen.Artifact{
			{Path: "src/index.ts", Content: []byte("console.log('hello');")},
			{Path: "README.md", Content: []byte("# readme")},
			{Path: "src/nested/deep/file.ts", Content: []byte("export {};")},
		},
	}
	err := stage.Run(ctx)
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(outDir, "src/index.ts"))
	assert.FileExists(t, filepath.Join(outDir, "README.md"))
	assert.FileExists(t, filepath.Join(outDir, "src/nested/deep/file.ts"))

	content, err := os.ReadFile(filepath.Join(outDir, "src/index.ts"))
	require.NoError(t, err)
	assert.Equal(t, "console.log('hello');", string(content))
}

func TestWriteStage_WriteOnceSkipsExisting(t *testing.T) {
	outDir := t.TempDir()

	// Pre-create a file that should be preserved
	existingPath := filepath.Join(outDir, "src", "impl.ts")
	require.NoError(t, os.MkdirAll(filepath.Dir(existingPath), 0755))
	require.NoError(t, os.WriteFile(existingPath, []byte("user code"), 0644))

	stage := Write()
	ctx := &Context{
		OutputDir: outDir,
		Artifacts: []codegen.Artifact{
			{Path: "src/impl.ts", Content: []byte("generated"), Strategy: codegen.WriteOnce},
		},
	}
	err := stage.Run(ctx)
	require.NoError(t, err)

	// File should still contain the original user code
	content, err := os.ReadFile(existingPath)
	require.NoError(t, err)
	assert.Equal(t, "user code", string(content))

	// Stats should show 1 skipped, 0 written
	assert.Equal(t, 0, ctx.Stats.Written)
	assert.Equal(t, 1, ctx.Stats.Skipped)
}

func TestWriteStage_WriteOnceCreatesNew(t *testing.T) {
	outDir := t.TempDir()

	stage := Write()
	ctx := &Context{
		OutputDir: outDir,
		Artifacts: []codegen.Artifact{
			{Path: "src/impl.ts", Content: []byte("generated"), Strategy: codegen.WriteOnce},
		},
	}
	err := stage.Run(ctx)
	require.NoError(t, err)

	// File should be created with generated content
	content, err := os.ReadFile(filepath.Join(outDir, "src/impl.ts"))
	require.NoError(t, err)
	assert.Equal(t, "generated", string(content))

	// Stats should show 1 written, 0 skipped
	assert.Equal(t, 1, ctx.Stats.Written)
	assert.Equal(t, 0, ctx.Stats.Skipped)
}

func TestWriteStage_WriteAlwaysOverwrites(t *testing.T) {
	outDir := t.TempDir()

	// Pre-create a file
	existingPath := filepath.Join(outDir, "src", "types.ts")
	require.NoError(t, os.MkdirAll(filepath.Dir(existingPath), 0755))
	require.NoError(t, os.WriteFile(existingPath, []byte("old content"), 0644))

	stage := Write()
	ctx := &Context{
		OutputDir: outDir,
		Artifacts: []codegen.Artifact{
			{Path: "src/types.ts", Content: []byte("new content"), Strategy: codegen.WriteAlways},
		},
	}
	err := stage.Run(ctx)
	require.NoError(t, err)

	// File should be overwritten
	content, err := os.ReadFile(existingPath)
	require.NoError(t, err)
	assert.Equal(t, "new content", string(content))

	// Stats should show 1 written, 0 skipped
	assert.Equal(t, 1, ctx.Stats.Written)
	assert.Equal(t, 0, ctx.Stats.Skipped)
}

func TestWriteStage_WriteStats(t *testing.T) {
	outDir := t.TempDir()

	// Pre-create one file to be skipped
	existingPath := filepath.Join(outDir, "src", "impl.ts")
	require.NoError(t, os.MkdirAll(filepath.Dir(existingPath), 0755))
	require.NoError(t, os.WriteFile(existingPath, []byte("user code"), 0644))

	stage := Write()
	ctx := &Context{
		OutputDir: outDir,
		Artifacts: []codegen.Artifact{
			{Path: "src/types.ts", Content: []byte("types"), Strategy: codegen.WriteAlways},
			{Path: "src/impl.ts", Content: []byte("generated"), Strategy: codegen.WriteOnce},
			{Path: "src/index.ts", Content: []byte("index"), Strategy: codegen.WriteAlways},
		},
	}
	err := stage.Run(ctx)
	require.NoError(t, err)

	assert.Equal(t, 2, ctx.Stats.Written)
	assert.Equal(t, 1, ctx.Stats.Skipped)
}

func TestFullValidationPipeline(t *testing.T) {
	p := New(
		Parse(),
		ValidateSchema(),
		BuildIR(),
		ValidateIR(),
	)

	ctx := &Context{SpecPath: "../../examples/basic/spec.yaml"}
	err := p.Run(ctx)

	require.NoError(t, err)
	assert.NotNil(t, ctx.AST)
	assert.NotNil(t, ctx.IR)
	assert.Greater(t, len(ctx.IR.Components), 0)
}
