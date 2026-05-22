package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/hayoung123/easesee/internal/config"
	"github.com/hayoung123/easesee/internal/registry"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List registered projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		paths := config.New()
		r, err := registry.Load(paths.RegistryFile)
		if err != nil {
			return err
		}
		if len(r.Projects) == 0 {
			fmt.Println("no projects registered")
			return nil
		}
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tCWD\tCMD")
		for _, p := range r.Projects {
			fmt.Fprintf(w, "%s\t%s\t%s\n", p.Name, p.Cwd, p.Cmd)
		}
		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
}
