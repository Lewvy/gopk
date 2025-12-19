package cmd

import (
	"github.com/lewvy/gopk/cmd/internal/service"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <alias> [alias...]",
	Short: "Install one or more saved packages into the current module",
	Long: `Install Go modules by alias from your gopk registry.

The get command resolves aliases stored in gopk and runs 'go get'
for each selected package in the current Go module.

This command is project-specific and requires an existing go.mod file.
It does not modify your gopk registry.`,

	Args: cobra.MinimumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		return service.Get(args, false, queries)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
