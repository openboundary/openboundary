// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package codegen

import (
	"fmt"
	"sort"
)

// Artifact represents a single planned output artifact.
type Artifact struct {
	Owner       string
	Path        string
	Content     []byte
	ComponentID string // The component that this artifact belongs to (empty for shared artifacts)
}

// ArtifactConflictError is returned when two generators write the same path.
type ArtifactConflictError struct {
	Path          string
	ExistingOwner string
	IncomingOwner string
}

func (e *ArtifactConflictError) Error() string {
	return fmt.Sprintf(
		"artifact path conflict for %q: already planned by %q, attempted by %q",
		e.Path, e.ExistingOwner, e.IncomingOwner,
	)
}

// ArtifactPlanner plans and deduplicates generated artifacts.
type ArtifactPlanner struct {
	byPath map[string]Artifact
}

// NewArtifactPlanner creates a new artifact planner.
func NewArtifactPlanner() *ArtifactPlanner {
	return &ArtifactPlanner{
		byPath: make(map[string]Artifact),
	}
}

// Add adds a single artifact to the plan.
func (p *ArtifactPlanner) Add(owner, path string, content []byte, componentID string) error {
	if path == "" {
		return fmt.Errorf("artifact path cannot be empty")
	}

	if existing, ok := p.byPath[path]; ok {
		return &ArtifactConflictError{
			Path:          path,
			ExistingOwner: existing.Owner,
			IncomingOwner: owner,
		}
	}

	artifactContent := make([]byte, len(content))
	copy(artifactContent, content)

	p.byPath[path] = Artifact{
		Owner:       owner,
		Path:        path,
		Content:     artifactContent,
		ComponentID: componentID,
	}

	return nil
}

// AddOutput adds a full generator output to the plan.
func (p *ArtifactPlanner) AddOutput(owner string, output *Output) error {
	if output == nil {
		return fmt.Errorf("generator %q returned nil output", owner)
	}

	paths := make([]string, 0, len(output.Files))
	for path := range output.Files {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		file := output.Files[path]
		if err := p.Add(owner, path, file.Content, file.ComponentID); err != nil {
			return err
		}
	}

	return nil
}

// Artifacts returns all planned artifacts sorted by path.
func (p *ArtifactPlanner) Artifacts() []Artifact {
	paths := make([]string, 0, len(p.byPath))
	for path := range p.byPath {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	artifacts := make([]Artifact, 0, len(paths))
	for _, path := range paths {
		artifacts = append(artifacts, p.byPath[path])
	}
	return artifacts
}
