package builds

import (
	"context"
	"fmt"
	"net/http"

	"github.com/evankanderson/knuts/pkg"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	// iamProto "google.golang.org/genproto/googleapis/iam/admin/v1"
	// iam "cloud.google.com/go/iam/admin/apiv1"
	iam "google.golang.org/api/iam/v1"
	serviceusage "google.golang.org/api/serviceusage/v1"
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
	client := oauth2.NewClient(ctx, creds.TokenSource)

	project := "projects" + creds.ProjectID
	// Step 1: ensure the correct services are enabled
	err = g.ensureServices(client, project, []string{"iam.googleapis.com", "containerregistry.googleapis.com"})
	if err != nil {
		return []byte{}
	}
	// Step 2: Create IAM Service account
	err = g.createServiceAccount(client, project)
	if err != nil {
		return []byte{}
	}

	//	iam.CreateServiceAccount(ctx, &iamProto.CreateServiceAccountRequest{Name: fmt.Sprintf("projects/%s", creds.ProjectID),
	//		AccountId: "push-image"})

	// Step 3: Assign new service account `roles.storage.admin` to all `*.artifacts.$project.artifacts.appspot.com` buckets

	// Step 4: Create a JSON Key for the Service Account

	return []byte{}
}

func (g GCPSecret) ensureServices(client *http.Client, project string, apis []string) error {
	smAPI, err := serviceusage.New(client)
	if err != nil {
		return err
	}
	// Check to see if we need to enable anything
	required := map[string]bool{}
	for _, a := range apis {
		required[a] = true
	}
	token := ""
	for len(required) > 0 {
		list, err := smAPI.Services.List(project).Filter("state:ENABLED").PageToken(token).Do()
		if err != nil {
			return err // TODO: should we just try to enable blindly?
		}
		for _, s := range list.Services {
			if required[s.Config.Name] {
				delete(required, s.Config.Name)
			}
		}
		token = list.NextPageToken
		if token == "" {
			break
		}
	}
	if len(required) == 0 {
		fmt.Printf("All services already enabled: %v\n", apis)
		return nil
	}
	apis = []string{}
	for api := range required {
		apis = append(apis, api)
	}
	if pkg.DryRun {
		fmt.Printf("Enabling APIs: %s", apis)
	}

	op, err := smAPI.Services.BatchEnable(
		project,
		&serviceusage.BatchEnableServicesRequest{ServiceIds: apis}).Do()
	if err != nil {
		return err
	}
	for !op.Done {
		op, err = smAPI.Operations.Get(op.Name).Do()
		if err != nil {
			return err
		}
	}
	return nil
}

func (g GCPSecret) createServiceAccount(client *http.Client, project string) error {
	iamAPI, err := iam.New(client)
	if err != nil {
		return err
	}
	if pkg.DryRun {
		fmt.Printf("Creating IAM account push-image in %s\n", project)
	}
	saService := iam.NewProjectsServiceAccountsService(iamAPI)
	sa, err := saService.Create(project,
		&iam.CreateServiceAccountRequest{
			AccountId: "push-image", 
			ServiceAccount: &iam.ServiceAccount{DisplayName: "Push images from cluster build"},
			}).Do()
	if err != nil {
		return err
	}
	fmt.Printf("Created ServiceAccount: %v", sa)
	return nil
}
