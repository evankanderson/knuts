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
			"bazel": {
				Description: "Bazel with container_push rule",
				Data:        "https://raw.githubusercontent.com/knative/build-templates/master/bazel/bazel.yaml",
			},
			"buildah": {
				Description: "Buildah mechanism for building from Dockerfiles. Requires $BUILDER_IMAGE set in your Build.",
				Data:        "https://raw.githubusercontent.com/knative/build-templates/master/buildah/buildah.yaml",
			},
			// TODO: buildkit is more complex, and may need additional cluster permissions.
		},
	}
)

// Install causes the BuildTemplate to be installed in the current kubernetes context.
func (f BuildTemplate) Install() error {
	return pkg.Kubectl(f.Data.(string))
}
