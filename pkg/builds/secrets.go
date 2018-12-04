package builds

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/evankanderson/knuts/pkg"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	// iamProto "google.golang.org/genproto/googleapis/iam/admin/v1"
	// iam "cloud.google.com/go/iam/admin/apiv1"
	iam "google.golang.org/api/iam/v1"
	serviceusage "google.golang.org/api/serviceusage/v1"
)

// ImageSecret contains the information needed to create an image push secret for build templates.
type ImageSecret struct {
	// Provider is a one word short description of the secret provider (used as the secret name).
	Provider string
	// Hosts describes the registry patterns which this secret applies to.
	Hosts []string
	// Username is the username used to authenticate to the registry.
	Username string
	// TODO: return a k8s secret object to be applied to the destination.
	Password string
}

// ProduceK8sSecret creates a kubernetes Secret objects suitable for application via kubectl.
func ProduceK8sSecret(s ImageSecret) ([]byte, error) {
	t, err := template.New("secret").Parse(`
apiVersion: v1
kind: Secret
metadata:
	name: {{ .Provider }}
	annotations:{{range $idx, $host := .Hosts}}
		build.knative.dev/docker-{{$idx}}: {{$host}}{{end}}
type: kubernetes.io/basic-auth
data:
	username: {{ .Username }}
	password: $(openssl base64 -a -A < image-push-key.json)
`)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	err = t.Execute(&b, s)
	return b.Bytes(), err
}

// Prompt will ask the user for credentials.
func Prompt() (ImageSecret, error) {
	fmt.Printf("This is where we would prompt for data")
	return ImageSecret{
		Provider: "prompted",
		Hosts:    []string{"docker.io"},
		Username: "prompted",
		Password: "",
	}, nil
}

// GCRSecret will create and grant permissions for a dedicated service account to call GCR.io.
func GCRSecret() (ImageSecret, error) {
	s, err := setupGCPSecret()
	return ImageSecret{
		Provider: "google-cloud-platform",
		Hosts:    []string{"us.gcr.io", "gcr.io", "eu.gcr.io", "asia.gcr.io"},
		Username: "X2pzb25fa2V5", // base64 encoded "_json_key"
		Password: s,
	}, err
}

// Secret is part of the Secret interface.
func setupGCPSecret() (string, error) {
	ctx := context.Background()
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", fmt.Errorf("Failed to fetch default credentials: %v", err)
	}
	if creds.ProjectID == "" {
		creds.ProjectID = pkg.GetGCPProject()
	}
	fmt.Printf("Using project %q to create services and enable registry\n", creds.ProjectID)
	client := oauth2.NewClient(ctx, creds.TokenSource)

	project := "projects/" + creds.ProjectID
	// Step 1: ensure the correct services are enabled
	err = ensureGCPServices(client, project, []string{"iam.googleapis.com", "containerregistry.googleapis.com"})
	if err != nil {
		return "", err
	}
	// Step 2: Create IAM Service account
	sa, err := createGCPServiceAccount(client, project)
	if err != nil {
		return "", err
	}
	fmt.Printf("Created ServiceAccount %q", sa.Email)

	//	iam.CreateServiceAccount(ctx, &iamProto.CreateServiceAccountRequest{Name: fmt.Sprintf("projects/%s", creds.ProjectID),
	//		AccountId: "push-image"})

	// Step 3: Assign new service account `roles.storage.admin` to all `*.artifacts.$project.artifacts.appspot.com` buckets

	// Step 4: Create a JSON Key for the Service Account

	return "", nil
}

func ensureGCPServices(client *http.Client, project string, apis []string) error {
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
		fmt.Printf("Enabling APIs: %s\n", apis)
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
	if op.Error != nil {
		return fmt.Errorf("Service enablement failed: %v", op.Error.Message)
	}
	return nil // TODO: check for op.Error
}

func createGCPServiceAccount(client *http.Client, project string) (*iam.ServiceAccount, error) {
	iamAPI, err := iam.New(client)
	if err != nil {
		return nil, err
	}
	shortProject := strings.TrimPrefix(project, "projects/")
	saName := "push-image"
	saEmail := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", saName, shortProject)
	if pkg.DryRun {
		fmt.Printf("Creating IAM account %q in %s\n", saName, project)
	}
	saService := iam.NewProjectsServiceAccountsService(iamAPI)
	existing, err := saService.Get(project + "/serviceAccounts/" + saEmail).Do()
	if err == nil && existing.Email == saEmail {
		return existing, nil
	}
	return saService.Create(project,
		&iam.CreateServiceAccountRequest{
			AccountId:      saName,
			ServiceAccount: &iam.ServiceAccount{DisplayName: "Push images from cluster build"},
		}).Do()
}
