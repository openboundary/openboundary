// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package pipeline

import (
	"github.com/openboundary/openboundary/internal/codegen"
	"github.com/openboundary/openboundary/internal/ir"
	"github.com/openboundary/openboundary/internal/parser"
)

// Context carries data between pipeline stages.
type Context struct {
	SpecPath  string
	OutputDir string
	AST       *parser.Spec
	IR        *ir.IR
	Artifacts []codegen.Artifact
}

// Stage is a single step in a pipeline.
type Stage interface {
	Name() string
	Run(ctx *Context) error
}

// Pipeline executes a sequence of stages.
type Pipeline struct {
	stages []Stage
}

// New creates a pipeline from the given stages.
func New(stages ...Stage) *Pipeline {
	return &Pipeline{stages: stages}
}

// Run executes each stage in order, stopping on the first error.
func (p *Pipeline) Run(ctx *Context) error {
	for _, s := range p.stages {
		if err := s.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}
