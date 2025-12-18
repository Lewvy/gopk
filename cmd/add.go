package cmd

import (
	"github.com/lewvy/gopk/cmd/internal/service"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <module-path>",
	Short: "Save a Go module for quick reuse",
	Long: `Add a Go module to your gopk registry.

The add command stores a module path under a human-friendly alias,
allowing you to quickly recall and install it in future projects.

By default, this command only records the module and does not modify
the current project. Use --install to immediately run 'go get' for
the added package in the current Go module.`,

	Args: cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]

		name, _ := cmd.Flags().GetString("name")
		version, _ := cmd.Flags().GetString("version")
		install, _ := cmd.Flags().GetBool("install")

		return service.Add(url, name, version, install, queries)
	},
}

func init() {
	addCmd.Flags().StringP("name", "n", "", "add package name")
	addCmd.Flags().StringP("version", "v", "latest", "add package version (used for go installs)")
	addCmd.Flags().BoolP("install", "i", false, "install the package")

	rootCmd.AddCommand(addCmd)

}
