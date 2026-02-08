// Copyright 2026 OpenBoundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package codegen

import "testing"

func TestArtifactPlanner_Add(t *testing.T) {
	p := NewArtifactPlanner()
	if err := p.Add("gen-a", "src/a.ts", []byte("a"), "comp-1", WriteAlways); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	artifacts := p.Artifacts()
	if len(artifacts) != 1 {
		t.Fatalf("Artifacts() len = %d, expected 1", len(artifacts))
	}
	if artifacts[0].Owner != "gen-a" {
		t.Errorf("owner = %q, expected %q", artifacts[0].Owner, "gen-a")
	}
	if artifacts[0].ComponentID != "comp-1" {
		t.Errorf("componentID = %q, expected %q", artifacts[0].ComponentID, "comp-1")
	}
}

func TestArtifactPlanner_Add_Conflict(t *testing.T) {
	p := NewArtifactPlanner()

	if err := p.Add("gen-a", "src/a.ts", []byte("a"), "comp-1", WriteAlways); err != nil {
		t.Fatalf("first Add() error = %v", err)
	}

	err := p.Add("gen-b", "src/a.ts", []byte("b"), "comp-2", WriteAlways)
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if _, ok := err.(*ArtifactConflictError); !ok {
		t.Fatalf("error type = %T, expected *ArtifactConflictError", err)
	}
}

func TestArtifactPlanner_Add_Strategy(t *testing.T) {
	p := NewArtifactPlanner()
	if err := p.Add("gen-a", "src/a.ts", []byte("a"), "comp-1", WriteOnce); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	artifacts := p.Artifacts()
	if len(artifacts) != 1 {
		t.Fatalf("Artifacts() len = %d, expected 1", len(artifacts))
	}
	if artifacts[0].Strategy != WriteOnce {
		t.Errorf("strategy = %v, expected WriteOnce", artifacts[0].Strategy)
	}
}

func TestArtifactPlanner_AddOutput_PropagatesStrategy(t *testing.T) {
	p := NewArtifactPlanner()
	output := NewOutput()
	output.AddFile("src/always.ts", []byte("always"))
	output.AddWriteOnceFile("src/once.ts", []byte("once"))

	if err := p.AddOutput("gen-a", output); err != nil {
		t.Fatalf("AddOutput() error = %v", err)
	}

	artifacts := p.Artifacts()
	if len(artifacts) != 2 {
		t.Fatalf("Artifacts() len = %d, expected 2", len(artifacts))
	}

	// Sorted by path: always.ts, once.ts
	if artifacts[0].Strategy != WriteAlways {
		t.Errorf("always.ts strategy = %v, expected WriteAlways", artifacts[0].Strategy)
	}
	if artifacts[1].Strategy != WriteOnce {
		t.Errorf("once.ts strategy = %v, expected WriteOnce", artifacts[1].Strategy)
	}
}

func TestArtifactPlanner_AddOutput(t *testing.T) {
	p := NewArtifactPlanner()
	output := NewOutput()
	output.AddFile("src/z.ts", []byte("z"))
	output.AddFile("src/a.ts", []byte("a"))

	if err := p.AddOutput("gen-a", output); err != nil {
		t.Fatalf("AddOutput() error = %v", err)
	}

	artifacts := p.Artifacts()
	if len(artifacts) != 2 {
		t.Fatalf("Artifacts() len = %d, expected 2", len(artifacts))
	}
	if artifacts[0].Path != "src/a.ts" || artifacts[1].Path != "src/z.ts" {
		t.Errorf("Artifacts() not sorted by path: %+v", artifacts)
	}
}
