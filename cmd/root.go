/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"

	"github.com/lewvy/gopk/cmd/internal/data"
	"github.com/lewvy/gopk/cmd/tui"
	"github.com/lewvy/gopk/config"
	"github.com/spf13/cobra"
)

var queries *data.Queries

var rootCmd = &cobra.Command{
	Use: "gopk",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		db, err := config.InitDB()
		if err != nil {
			log.Fatalf("error initializing db: %q", err)
		}

		queries = data.New(db)

	},
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Start(queries)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
