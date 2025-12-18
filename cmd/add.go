package cmd

import (
	"github.com/lewvy/gopk/cmd/internal/service"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add <url> [flags]",
	Short: "add a package",
	Long:  `add a package and store it in the config directory`,

	Args: cobra.RangeArgs(1, 1),

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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
