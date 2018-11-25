package builds

// BuildTemplate represents the steps needed to install a particular BuildTemplate.
type BuildTemplate struct {
	// Short is a single-word command-line flag name for the BuildTemplate.
	Short string
	// Desc contains a human-readable short description of the BuildTemplate.
	Desc string
	// URL of the current (YAML) definition of this template.
	URL string
	// Required secrets for this template.
	Secrets []SecretType
}

// Builds contains the set of known BuildTemplates.
var Builds = []BuildTemplate{
	{
		Short:   "jib-gradle",
		Desc:    "Gradle build with JIB",
		URL:     "https://github.com/knative/build-templates/blob/master/jib/jib-gradle.yaml",
		Secrets: []SecretType{ImagePush},
	},
	{
		Short:   "jib-maven",
		Desc:    "Maven build with JIB",
		URL:     "https://github.com/knative/build-templates/blob/master/jib/jib-maven.yaml",
		Secrets: []SecretType{ImagePush},
	},
	{
		Short:   "kaniko",
		Desc:    "Dockerfile with Kaniko",
		URL:     "https://github.com/knative/build-templates/blob/master/kaniko/kaniko.yaml",
		Secrets: []SecretType{ImagePush},
	},
	{
		Short:   "buildpack",
		Desc:    "Buildpack",
		URL:     "https://github.com/knative/build-templates/blob/master/buildpack/buildpack.yaml",
		Secrets: []SecretType{ImagePush},
	},
}
