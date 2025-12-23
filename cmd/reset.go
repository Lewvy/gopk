/*
Copyright © 2025 [Your Name] <[Your Email]>
*/
package cmd

import (
	"fmt"

	"github.com/lewvy/gopk/cmd/internal/service"
	"github.com/spf13/cobra"
)

// resetCmd represents the reset command
var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Permanently purge all packages",
	Long: `The reset command performs a 'hard delete' on all packages and groups. 

WARNING: This action is irreversible. If you are using multi-device sync, 
ensure all other devices have performed a sync BEFORE running this command, 
otherwise deleted items may reappear (resurrect) during the next sync.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		err := service.Reset()
		if err != nil {
			return fmt.Errorf("failed to purge deleted records: %w", err)
		}

		fmt.Println("Cleanup complete: All soft-deleted records have been permanently removed.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)

	// resetCmd.Flags().BoolP("force", "f", false, "Force purge without confirmation")
}
