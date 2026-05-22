package cli

import (
	"github.com/spf13/cobra"

	"github.com/proshy/devs/internal/tui"
)

var rootCmd = &cobra.Command{
	Use:   "devs",
	Short: "devs — local dev server dashboard",
	Long:  "A TUI for managing registered local dev servers. Run with no args to launch the dashboard.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run()
	},
}

func Execute() error {
	return rootCmd.Execute()
}
