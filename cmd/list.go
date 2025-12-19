package cmd

import (
	"fmt"

	"github.com/lewvy/gopk/cmd/internal/service"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved packages",
	Long: `List all packages saved in your gopk registry.
    
By default, packages are sorted by when they were last used.
Use the --freq flag to sort by most frequently used instead.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		byFreq, _ := cmd.Flags().GetBool("freq")

		pkgs, err := service.List(queries, limit, byFreq)
		for _, p := range pkgs {
			fmt.Println(p.Name, p.Url, p.Freq.Int64)
		}

		return err
	},
}

func init() {
	listCmd.Flags().IntP("limit", "l", -1, "limit the number of results")
	listCmd.Flags().BoolP("freq", "f", false, "sort results by frequency of use")

	rootCmd.AddCommand(listCmd)
}
