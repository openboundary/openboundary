// Copyright 2026 Open Boundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit_BlankTemplate(t *testing.T) {
	dir := t.TempDir()
	project := filepath.Join(dir, "my-blank-project")

	err := initInDir(dir, "my-blank-project", "blank")
	require.NoError(t, err)

	// spec.yaml should exist
	specPath := filepath.Join(project, "spec.yaml")
	assert.FileExists(t, specPath)

	content, err := os.ReadFile(specPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "name: my-blank-project")
}

func TestInit_BasicTemplate(t *testing.T) {
	dir := t.TempDir()
	project := filepath.Join(dir, "my-basic-project")

	err := initInDir(dir, "my-basic-project", "basic")
	require.NoError(t, err)

	expectedFiles := []string{
		"spec.yaml",
		"config/better-auth.config.ts",
		"config/casbin.model.conf",
		"config/casbin.policy.csv",
		"config/drizzle.schema.ts",
		"config/openapi.schema.yaml",
	}
	for _, f := range expectedFiles {
		assert.FileExists(t, filepath.Join(project, f), "missing file: %s", f)
	}
}

func TestInit_ProjectNameSubstitution(t *testing.T) {
	dir := t.TempDir()

	err := initInDir(dir, "cool-api", "basic")
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "cool-api", "spec.yaml"))
	require.NoError(t, err)

	assert.Contains(t, string(content), "name: cool-api")
	assert.NotContains(t, string(content), "name: user-api")
}

func TestInit_DirectoryAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "existing-project")
	require.NoError(t, os.Mkdir(existing, 0755))

	err := initInDir(dir, "existing-project", "blank")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestInit_InvalidProjectName(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
	}{
		{"absolute path", "/tmp/evil"},
		{"parent traversal", "../../escape"},
		{"forward slash", "path/to/project"},
		{"backslash", `path\to\project`},
		{"dot-dot", ".."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Init(tt.projectName, "blank")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid project name")
		})
	}
}

func TestInit_UnknownTemplate(t *testing.T) {
	dir := t.TempDir()

	err := initInDir(dir, "my-project", "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown template")
}

func TestInit_BasicTemplateValidates(t *testing.T) {
	dir := t.TempDir()

	err := initInDir(dir, "test-project", "basic")
	require.NoError(t, err)

	specPath := filepath.Join(dir, "test-project", "spec.yaml")
	err = Validate(specPath)
	assert.NoError(t, err)
}

func TestInit_BlankTemplateValidates(t *testing.T) {
	dir := t.TempDir()

	err := initInDir(dir, "test-project", "blank")
	require.NoError(t, err)

	specPath := filepath.Join(dir, "test-project", "spec.yaml")
	err = Validate(specPath)
	assert.NoError(t, err)
}

// initInDir runs Init from within the given directory so projects are created
// inside the temp dir rather than the working directory.
func initInDir(dir, projectName, template string) error {
	orig, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(dir); err != nil {
		return err
	}
	defer os.Chdir(orig) //nolint:errcheck // best-effort restore
	return Init(projectName, template)
}
