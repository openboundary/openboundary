// Copyright 2026 OpenBoundary Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/openboundary/openboundary/internal/codegen/typescript"
	"github.com/openboundary/openboundary/internal/pipeline"
)

func Compile(specFile string, outputDir string) error {
	p := pipeline.New(
		pipeline.Parse(),
		pipeline.ValidateSchema(),
		pipeline.BuildIR(),
		pipeline.ValidateIR(),
		pipeline.Generate(typescript.NewPluginRegistry),
		pipeline.Write(),
	)

	ctx := &pipeline.Context{
		SpecPath:  specFile,
		OutputDir: outputDir,
	}

	if err := p.Run(ctx); err != nil {
		printStageError(err)
		return err
	}

	fmt.Printf("\n✓ Generated %d files (%d written, %d preserved) in %s/\n",
		len(ctx.Artifacts), ctx.Stats.Written, ctx.Stats.Skipped, outputDir)
	return nil
}

func printStageError(err error) {
	var stageErr *pipeline.StageError
	if errors.As(err, &stageErr) {
		fmt.Fprintf(os.Stderr, "%s with %d error(s):\n", stageErr.Message, len(stageErr.Errors))
		for _, e := range stageErr.Errors {
			fmt.Fprintf(os.Stderr, "  - %s\n", e.Error())
		}
	}
}
