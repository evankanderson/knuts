package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/evankanderson/knuts/pkg"
	"github.com/evankanderson/knuts/pkg/builds"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildTemplateCmd)
	buildTemplateCmd.PersistentFlags().Var(&builds.Builds, "templates", builds.Builds.Description)
	buildTemplateCmd.PersistentFlags().Var(&gcpProject, "gcp_project", gcpProject.Description)
	buildTemplateCmd.PersistentFlags().Var(&registries, "registries", registries.Description)
}

var (
	registries = pkg.MultiSelect{
		Description: "Which registries to push to",
		Options: map[string]pkg.Option{
			"docker": {
				Description: "Docker (user secret)",
				Data:        builds.Prompt,
			},
			"gcr.io": {
				Description: "Google Container Registry",
				Data: func() (builds.ImageSecret, error) {
					return builds.GCRSecret(gcpProject.Get().(string))
				},
			},
		},
	}
	gcpProject = pkg.Prompt{
		Description: "GCP Project to push images to",
	}
)

var buildTemplateCmd = &cobra.Command{
	Use:     "build_templates",
	Aliases: []string{"builds", "bt"},
	Short:   "Menu-guided install of build templates.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := pkg.Installed("kubectl"); err != nil {
			fmt.Print(err)
			os.Exit(2)
		}

		templates := builds.Builds.Get().([]pkg.Option)

		for _, t := range templates {
			b := builds.BuildTemplate(t)
			err := b.Install()
			
			if err != nil {
				if ee, ok := err.(*exec.ExitError); ok {
					fmt.Printf("Failed to install %s: %v:\n%s\n", b.Description, ee, ee.Stderr)
				} else {
					fmt.Printf("Failed to install %s: %v\n", b.Description, err)
				}
				// For now, continue to the next install
			}
		}

		if len(templates) > 0 {
			// Set up registry secrets
			secrets := []builds.ImageSecret{}
			for _, f := range registries.Get().([]pkg.Option) {
				setup := f.Data.(func() (builds.ImageSecret, error))
				s, err := setup()
				if err != nil {
					fmt.Printf("Failed to set up %q: %v", f.Description, err)
				}
				secrets = append(secrets, s)
			}

			for _, s := range secrets {
				out, err := builds.ProduceK8sSecret(s)
				if err != nil {
					fmt.Printf("Skipping secret %q: %v", s.Provider, err)
					continue
				}
				fmt.Printf("%s\n", out)
			}
		}
	},
}