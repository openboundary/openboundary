package commands

import "github.com/spf13/cobra"

func Init(cmd *cobra.Command, args []string) {

	template := args[0]
	if template == "" {
		template = "blank"
	}

	// TODO: Bootstrap application
	//
	// Options:
	// - Blank
	// - Basic

}
