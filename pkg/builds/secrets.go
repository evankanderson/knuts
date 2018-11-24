package builds

import (
	"fmt"
)

// Secret represents a process for requesting configuration of a secret
type Secret interface {
	// Provider is a one short description of the secret provider
	Provider() string
	// TODO: return a k8s secret object to be applied to the destination.
	Secret() []byte
}

// SecretType is an enum representing the different types of Secrets which a
// BuildTemplate might need.
type SecretType int

const (
	// ImagePush secrets contain credentials for talking to an image registry.
	ImagePush SecretType = iota
	// GitPull secrets contain credentials for interacting with a remote git repo.
	GitPull
)

type imageProviders map[SecretType][]Secret

// KnownSecrets contains the known providers for different secret types.
var KnownSecrets = imageProviders{
	ImagePush: {prompt{}},
	GitPull:   {prompt{}},
}

type prompt struct {
	content string
}

func (prompt) Provider() string {
	return "static string"
}

func (prompt) Secret() []byte {
	fmt.Print("This would ask for a string and stuff it into a secret.")
	return []byte{}
}
