package cmd

import (
	"context"
	"fmt"

	"github.com/lewvy/gopk/cmd/internal/data"
	"github.com/lewvy/gopk/cmd/internal/service"
	"github.com/spf13/cobra"
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove packages or entire groups from your local database",
	Long: `The rm command allows you to delete specific packages or an entire group of packages.
By default, this performs a soft-delete to maintain sync compatibility.

Examples:
  gopk rm -n my-package
  gopk rm -n pkg1,pkg2,pkg3
  gopk rm -g my-project-group`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pkgs, _ := cmd.Flags().GetStringSlice("names")
		g, _ := cmd.Flags().GetString("group")

		ctx := context.Background()

		// 1. Handle Group Deletion
		if g != "" {
			err := service.DeleteGroup(ctx, queries, data.Group{Name: g})
			if err != nil {
				return fmt.Errorf("failed to delete group %q: %w", g, err)
			}
			fmt.Printf("Successfully marked group %q as deleted\n", g)
		}

		// 2. Handle Individual Package Deletion
		if len(pkgs) > 0 {
			err := service.DeletePackage(ctx, queries, pkgs)
			if err != nil {
				return fmt.Errorf("failed to delete packages: %w", err)
			}
			fmt.Printf("Successfully marked %d package(s) as deleted\n", len(pkgs))
		}

		// Validation: If no flags were provided
		if g == "" && len(pkgs) == 0 {
			return fmt.Errorf("must specify at least one package (-n) or a group (-g)")
		}

		return nil
	},
}

func init() {
	rmCmd.Flags().StringP("group", "g", "", "name of the group to delete")
	rmCmd.Flags().StringSliceP("names", "n", []string{}, "list the packages to delete")
	rootCmd.AddCommand(rmCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// rmCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// rmCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
