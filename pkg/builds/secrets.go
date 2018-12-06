package builds

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"google.golang.org/api/storage/v1"

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
	// Password is string-formatted private authentication data.
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
  password: {{ .Password }}
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
func GCRSecret(project string) (ImageSecret, error) {
	s, err := setupGCPSecret(project)
	return ImageSecret{
		Provider: "google-cloud-platform",
		Hosts:    []string{"us.gcr.io", "gcr.io", "eu.gcr.io", "asia.gcr.io"},
		Username: "X2pzb25fa2V5", // base64 encoded "_json_key"
		Password: s,
	}, err
}

// Secret is part of the Secret interface.
func setupGCPSecret(project string) (string, error) {
	ctx := context.Background()
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", fmt.Errorf("Failed to fetch default credentials: %v", err)
	}
	if creds.ProjectID == "" {
		creds.ProjectID = project
	}
	fmt.Printf("Using project %q to create services and enable registry\n", creds.ProjectID)
	client := oauth2.NewClient(ctx, creds.TokenSource)

	project = "projects/" + creds.ProjectID
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
	fmt.Printf("Created ServiceAccount %q\n", sa.Email)

	// Step 3: Assign new service account `roles.storage.admin` to all `*.artifacts.$project.artifacts.appspot.com` buckets
	for _, region := range []string{"artifacts", "us.artifacts", "eu.artifacts", "asia.artifacts"} {
		err := setIamPermissions(client, creds.ProjectID, region, sa)
		if err != nil {
			fmt.Printf("Failed to set permissions in %s: %v\n", region, err)
		}
	}

	// Step 4: Create a JSON Key for the Service Account
	return getSAKey(client, sa)
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
	saService := iam.NewProjectsServiceAccountsService(iamAPI)
	existing, err := saService.Get(project + "/serviceAccounts/" + saEmail).Do()
	if err == nil && existing.Email == saEmail {
		return existing, nil
	}
	if pkg.DryRun {
		fmt.Printf("Creating IAM account %q in %s\n", saName, project)
		return &iam.ServiceAccount{
			Email:    saEmail,
			UniqueId: "1234",
		}, nil
	}
	return saService.Create(project,
		&iam.CreateServiceAccountRequest{
			AccountId:      saName,
			ServiceAccount: &iam.ServiceAccount{DisplayName: "Push images from cluster build"},
		}).Do()
}

func setIamPermissions(client *http.Client, project string, region string, serviceAccount *iam.ServiceAccount) error {
	storageAPI, err := storage.New(client)
	if err != nil {
		return err
	}
	bService := storage.NewBucketsService(storageAPI)
	bucketName := fmt.Sprintf("%s.%s.appspot.com", region, project)
	p, err := bService.GetIamPolicy(bucketName).Do()
	if err != nil {
		return err
	}
	if p == nil {
		p = &storage.Policy{}
	}
	var newBinding *storage.PolicyBindings
	for _, binding := range p.Bindings {
		if binding.Role == "roles/storage.admin" {
			newBinding = binding
			break
		}
	}
	if newBinding == nil {
		newBinding = &storage.PolicyBindings{
			Role:    "roles/storage.admin",
			Members: []string{},
		}
		p.Bindings = append(p.Bindings, newBinding)
	}
	newMember := "serviceAccount:" + serviceAccount.Email
	for _, m := range newBinding.Members {
		if m == newMember {
			newMember = "" // already present
			break
		}
	}
	if newMember != "" {
		newBinding.Members = append(newBinding.Members, newMember)
	}
	if pkg.DryRun {
		fmt.Printf("Would add %s to %s\n", newMember, bucketName)
		return nil
	}
	_, err = bService.SetIamPolicy(bucketName, p).Do()
	return err
}

func getSAKey(client *http.Client, sa *iam.ServiceAccount) (string, error) {
	iamAPI, err := iam.New(client)
	if err != nil {
		return "", err
	}
	keyService := iam.NewProjectsServiceAccountsKeysService(iamAPI)
	if pkg.DryRun {
		return "FAKE", nil
	}
	key, err := keyService.Create("projects/-/serviceAccounts/"+sa.UniqueId, &iam.CreateServiceAccountKeyRequest{}).Do()
	if err != nil {
		return "", err
	}
	return key.PrivateKeyData, nil
}
