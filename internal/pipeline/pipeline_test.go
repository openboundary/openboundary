// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package pipeline

import (
	"errors"
	"testing"

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
