package builds

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	servicemanagement "google.golang.org/api/servicemanagement/v1"
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

// GCPSecret contains the data needed to return a GCP image push secret
type GCPSecret struct {
	serviceAccount string
	refreshToken   string
}

// Provider is part of the Secret interface.
func (g GCPSecret) Provider() string {
	return "google-cloud-platform"
}

// Secret is part of the Secret interface.
func (g GCPSecret) Secret() []byte {
	ctx := context.Background()
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		fmt.Printf("Failed to fetch default credentials: %v", err)
		return []byte{} // TODO: error handling
	}
	fmt.Printf("Using project %s to create services and enable registry", creds.ProjectID)
	http := oauth2.NewClient(ctx, creds.TokenSource)

	smAPI, err := servicemanagement.New(http)
	if err != nil {
		return []byte{}
	}
	smService := servicemanagement.NewServicesService(smAPI)
	opService := servicemanagement.NewOperationsService(smAPI)
	// Step 1: ensure the correct services are enabled
	ops := []*servicemanagement.Operation{}
	for _, s := range []string{"iam", "containerregistry"} {
		op, err := smService.Enable(s, &servicemanagement.EnableServiceRequest{ConsumerId: fmt.Sprintf(creds.ProjectID)}).Do()
		if err != nil {
			return []byte{}
		}
		ops = append(ops, op)
	}
	for len(ops) > 0 {
		for i, op := range ops {
			op, err := opService.Get(op.Name).Do()
			if err != nil {
				return []byte{}
			}
			if op.Done {
				ops = append(ops[:i], ops[i+1:]...)
			}
		}
	}
	// Step 2: Create IAM Service account

	// Step 3: Assign new service account `roles.storage.admin` to all `*.artifacts.$project.artifacts.appspot.com` buckets

	// Step 4: Create a JSON Key for the Service Account

	return []byte{}
}
