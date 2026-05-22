package cli

import (
	"github.com/spf13/cobra"

	"github.com/hayoung123/easesee/internal/tui"
)

// Version is set at build time via ldflags (-X cli.Version=...).
// Falls back to "dev" for unversioned local builds.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "easesee",
	Short:   "easesee — local dev server dashboard",
	Long:    "A TUI for managing registered local dev servers. Run with no args to launch the dashboard.",
	Version: Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run()
	},
}

func Execute() error {
	rootCmd.Version = Version
	return rootCmd.Execute()
}
