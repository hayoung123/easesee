package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "devs",
	Short: "devs — local dev server dashboard",
	Long:  "A TUI for managing registered local dev servers. Run with no args to launch the dashboard.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TUI launch wired up in a later task.
		cmd.Println("(TUI not yet implemented — use `devs register --help`)")
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}
