package commands

import (
	"fmt"

	"github.com/openboundary/openboundary/internal/pipeline"
)

func Validate(specFile string) error {
	p := pipeline.New(
		pipeline.Parse(),
		pipeline.ValidateSchema(),
		pipeline.BuildIR(),
		pipeline.ValidateIR(),
	)

	ctx := &pipeline.Context{SpecPath: specFile}

	if err := p.Run(ctx); err != nil {
		printStageError(err)
		return err
	}

	fmt.Printf("✓ %s is valid (version: %s, name: %s, %d components)\n",
		specFile, ctx.AST.Version, ctx.AST.Name, len(ctx.AST.Components))
	return nil
}
