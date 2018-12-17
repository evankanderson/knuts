package cmd

import (
	"fmt"
	"os"

	"github.com/evankanderson/knuts/pkg/install"

	"github.com/evankanderson/knuts/pkg"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.PersistentFlags().Var(&componentsFlag, "components", componentsFlag.Description)
}

var (
	componentsFlag = install.ComponentsAsFlag()
)

var installCmd = &cobra.Command{
	Use:     "install",
	Aliases: []string{"in", "knstall"},
	Short:   "Menu-guided install of Knative components.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := pkg.Installed("kubectl"); err != nil {
			fmt.Print(err)
			os.Exit(2)
		}
		selected := componentsFlag.Get().([]pkg.Option)
		work := []install.Component{}
		for _, o := range selected {
			component := o.Data.(install.Component)
			work = component.Expand(work)
		}
		for _, w := range work {
			pkg.Kubectl(w.Yaml, os.Stdout)
		}
	},
}
