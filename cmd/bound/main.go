// Copyright 2026 OpenBoundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package main provides the CLI entry point for the openboundary compiler.
package main

import (
	"fmt"
	"os"

	"github.com/openboundary/openboundary/cmd/bound/commands"
	"github.com/spf13/cobra"
)

var (
	version          = "0.1.0"
	compileOutputDir string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "bound",
		Short: "OpenBoundary specification compiler",
		Long: `bound compiles executable specifications into runnable code.

		It reads YAML specification files and generates type-safe code
		for various target platforms.`,
	}

	// Version flag
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("bound version {{.Version}}\n")

	// init command
	var initTemplate string
	initCmd := &cobra.Command{
		Use:   "init <project-name>",
		Short: "Initialize a new project from a template",
		Long:  `Initialize a new project directory from a template (blank or basic).`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.Init(args[0], initTemplate)
		},
	}
	initCmd.Flags().StringVarP(&initTemplate, "template", "t", "blank", "Template to use (blank, basic)")

	// validate command
	validateCmd := &cobra.Command{
		Use:   "validate [spec-file]",
		Short: "Validate a specification file",
		Long:  `Validate a specification file against the OpenBoundary schema and semantic rules.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.Validate(args[0])
		},
	}

	// compile command
	compileCmd := &cobra.Command{
		Use:   "compile [spec-file]",
		Short: "Compile a specification file",
		Long:  `Compile a specification file into executable code for the target platform.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.Compile(args[0], compileOutputDir)
		},
	}
	compileCmd.Flags().StringVarP(&compileOutputDir, "output", "o", "generated", "Output directory for generated code")

	rootCmd.AddCommand(compileCmd, validateCmd, initCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
