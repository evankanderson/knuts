package cmd

import (
	"fmt"
	"os"

	"github.com/evankanderson/knuts/pkg"
	"github.com/evankanderson/knuts/pkg/builds"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildTemplateCmd)
}

var buildTemplateCmd = &cobra.Command{
	Use:     "build_templates",
	Aliases: []string{"builds", "bt"},
	Short:   "Menu-guided install of build templates.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := pkg.Installed("kubectl"); err != nil {
			fmt.Print(err)
			os.Exit(2)
		}

		fmt.Println("Build template options:")
		for _, b := range builds.Builds {
			fmt.Printf("  %s\n", b.Desc)
		}
	},
}
