package cmd

import (
	"os/exec"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/evankanderson/knuts/pkg"
	"github.com/evankanderson/knuts/pkg/builds"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildTemplateCmd)
	selected = buildTemplateCmd.PersistentFlags().StringSlice(
		"templates", []string{}, "Names of build templates to install. If not provided, will prompt interactively.")
}

var byName = prepare(builds.Builds)

var selected *[]string

var buildTemplateCmd = &cobra.Command{
	Use:     "build_templates",
	Aliases: []string{"builds", "bt"},
	Short:   "Menu-guided install of build templates.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(*selected) == 0 {
			selected = fromPrompt(byName)
		}
		if err := pkg.Installed("kubectl"); err != nil {
			fmt.Print(err)
			os.Exit(2)
		}

		installs := []builds.BuildTemplate{}
		for _, n := range *selected {
			b := byName[n]
			installs = append(installs, b)
			err := b.Install()
			if err != nil {
				if ee, ok := err.(*exec.ExitError); ok {
					fmt.Printf("Failed to install %s (%s): %v:\n%s\n", b.Short, b.Desc, ee, ee.Stderr)
				} else {
				fmt.Printf("Failed to install %s (%s): %v\n", b.Short, b.Desc, err)
				}
				// For now, continue to the next install
			}
		}
		if len(installs) > 0 {
			// Prompt for secret.
			secretOpts := &survey.MultiSelect{
				Message: "Set up image push for which registries?",
				Options: []string{"Manual prompt", "GCR.io"},
			}
			answers := []string{}
			err := survey.AskOne(secretOpts, &answers, nil)
			if err != nil {
				fmt.Printf("Failed reading registration: %v", err)
				return
			}
			secrets := []builds.ImageSecret{}
			for _, a := range answers {
				switch a {
				case "Manual prompt":
					s, err := builds.Prompt()
					if err != nil {
						fmt.Printf("Failed to generated prompted Secret: %v.", err)
						continue
					}
					secrets = append(secrets, s)
				case "GCR.io":
					s, err := builds.GCRSecret()
					if err != nil {
						fmt.Printf("Failed to generate GCR Secret: %v", err)
						continue
					}
					secrets = append(secrets, s)
				}
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

func prepare(templates []builds.BuildTemplate) map[string]builds.BuildTemplate {
	byName := map[string]builds.BuildTemplate{}
	for _, b := range builds.Builds {
		byName[b.Short] = b
	}
	return byName
}

func fromPrompt(m map[string]builds.BuildTemplate) *[]string {
	buildChoices := make([]string, len(byName))
	i := 0
	for _, v := range byName {
		buildChoices[i] = fmt.Sprintf("%s: %s", v.Short, v.Desc)
		i++
	}
	sort.Strings(buildChoices)
	prompt := &survey.MultiSelect{
		Message: "Which BuildTemplates do you want to install?",
		Options: buildChoices,
	}
	answers := []string{}
	err := survey.AskOne(prompt, &answers, nil)
	if err != nil {
		fmt.Printf("Error prompting for builds: %s\n", err)
		os.Exit(1)
	}
	ret := make([]string, len(answers))
	for i, s := range answers {
		ret[i] = strings.SplitN(s, ":", 2)[0]
	}
	return &ret
}
