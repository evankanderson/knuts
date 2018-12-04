package pkg

import (
	"github.com/AlecAivazis/survey"
)

var DryRun = true
var GCPProject = ""

func GetGCPProject() string {
	if GCPProject != "" {
		return GCPProject
	}
	prompt := &survey.Input{Message: "Enter your GCP Project:"}
	survey.AskOne(prompt, &GCPProject, nil)
	return GCPProject
}
