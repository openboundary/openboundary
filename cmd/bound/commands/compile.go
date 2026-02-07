package commands

import (
	"fmt"

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
		return err
	}

	fmt.Printf("\n✓ Generated %d files in %s/\n", len(ctx.Artifacts), outputDir)
	return nil
}
