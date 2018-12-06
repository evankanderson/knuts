package builds

import (
	"github.com/evankanderson/knuts/pkg"
)

// BuildTemplate represents the steps needed to install a particular BuildTemplate.
type BuildTemplate pkg.Option

var (
	// Builds contains the set of known BuildTemplates.
	Builds = pkg.MultiSelect{
		Description: "Which build templates do you want to install",
		Options: map[string]pkg.Option{
			"jib-gradle": {
				Description: "Gradle build with JIB",
				Data:        "https://raw.githubusercontent.com/knative/build-templates/master/jib/jib-gradle.yaml",
			},
			"jib-maven": {
				Description: "Maven build with JIB",
				Data:        "https://raw.githubusercontent.com/knative/build-templates/master/jib/jib-maven.yaml",
			},
			"kaniko": {
				Description: "Dockerfile with Kaniko",
				Data:        "https://raw.githubusercontent.com/knative/build-templates/master/kaniko/kaniko.yaml",
			},
			"buildpack": {
				Description: "Buildpack",
				Data:        "https://raw.githubusercontent.com/knative/build-templates/master/buildpack/buildpack.yaml",
			},
		},
	}
)

// Install causes the BuildTemplate to be installed in the current kubernetes context.
func (f BuildTemplate) Install() error {
	return pkg.Kubectl(f.Data.(string))
}
