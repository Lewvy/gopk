package cmd

import (
	"github.com/lewvy/gopk/cmd/internal/service"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add <url> [alias]",
	Short: "add a package",
	Long:  `add a package and store it in the config directory`,
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {

		c := cmd.Flag("n")
		var alias string
		if c != nil {
			alias = c.Value.String()
			args = append(args, alias)
		}
		return service.Add(args, queries)
	},
}

func init() {
	addCmd.Flags().String("n", "", "alias for a url")
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
