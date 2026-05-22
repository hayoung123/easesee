package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proshy/easesee/internal/config"
	"github.com/proshy/easesee/internal/registry"
)

var registerFlags struct {
	Name  string
	Cwd   string
	Cmd   string
	Force bool
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a project in the devs registry",
	RunE: func(cmd *cobra.Command, args []string) error {
		paths := config.New()
		if err := paths.EnsureDirs(); err != nil {
			return err
		}
		return registerProject(paths.RegistryFile, registry.Project{
			Name: registerFlags.Name,
			Cwd:  registerFlags.Cwd,
			Cmd:  registerFlags.Cmd,
		}, registerFlags.Force)
	},
}

func init() {
	registerCmd.Flags().StringVar(&registerFlags.Name, "name", "", "project name (required)")
	registerCmd.Flags().StringVar(&registerFlags.Cwd, "cwd", "", "working directory (required)")
	registerCmd.Flags().StringVar(&registerFlags.Cmd, "cmd", "", "command to start dev server (required)")
	registerCmd.Flags().BoolVar(&registerFlags.Force, "force", false, "replace if name already exists")
	_ = registerCmd.MarkFlagRequired("name")
	_ = registerCmd.MarkFlagRequired("cwd")
	_ = registerCmd.MarkFlagRequired("cmd")
	rootCmd.AddCommand(registerCmd)
}

func registerProject(registryPath string, p registry.Project, force bool) error {
	cwd, err := expandPath(p.Cwd)
	if err != nil {
		return err
	}
	if _, err := os.Stat(cwd); err != nil {
		return fmt.Errorf("cwd does not exist: %s", cwd)
	}
	p.Cwd = cwd

	r, err := registry.Load(registryPath)
	if err != nil {
		return err
	}
	if force {
		if _, ok := r.Find(p.Name); ok {
			if err := r.Replace(p); err != nil {
				return err
			}
		} else if err := r.Add(p); err != nil {
			return err
		}
	} else {
		if err := r.Add(p); err != nil {
			return err
		}
	}
	if err := r.Save(registryPath); err != nil {
		return err
	}
	fmt.Printf("registered %q (cwd=%s, cmd=%q)\n", p.Name, p.Cwd, p.Cmd)
	return nil
}

func expandPath(p string) (string, error) {
	if strings.HasPrefix(p, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		p = filepath.Join(home, strings.TrimPrefix(p, "~"))
	}
	return filepath.Abs(p)
}
