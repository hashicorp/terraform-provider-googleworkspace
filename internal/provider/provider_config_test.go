package googleworkspace

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	googleoauth "golang.org/x/oauth2/google"

	"google.golang.org/api/iamcredentials/v1"
	"google.golang.org/api/option"
)

func TestConfigLoadAndValidate_credsInvalidJSON(t *testing.T) {
	config := &apiClient{
		Credentials:           "{this is not json}",
		ImpersonatedUserEmail: "my-fake-email@example.com",
	}

	diags := config.loadAndValidate(context.Background())
	if !diags.HasError() {
		t.Fatalf("expected error, but got nil")
	}
}

func TestConfigLoadAndValidate_credsJSON(t *testing.T) {
	contents, err := ioutil.ReadFile(testFakeCredentialsPath)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	config := &apiClient{
		Credentials:           string(contents),
		ImpersonatedUserEmail: "my-fake-email@example.com",
	}

	diags := config.loadAndValidate(context.Background())
	err = checkDiags(diags)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestConfigLoadAndValidate_credsFromFile(t *testing.T) {
	config := &apiClient{
		Credentials:           testFakeCredentialsPath,
		ImpersonatedUserEmail: "my-fake-email@example.com",
	}

	diags := config.loadAndValidate(context.Background())
	err := checkDiags(diags)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestAccConfigLoadAndValidate_credsFromEnv(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip(fmt.Sprintf("Network access not allowed; use TF_ACC=1 to enable"))
	}

	testAccPreCheck(t)

	creds := getTestCredsFromEnv()
	config := &apiClient{
		Credentials:           creds,
		Customer:              os.Getenv("GOOGLEWORKSPACE_CUSTOMER_ID"),
		ImpersonatedUserEmail: os.Getenv("GOOGLEWORKSPACE_IMPERSONATED_USER_EMAIL"),
	}

	diags := config.loadAndValidate(context.Background())
	err := checkDiags(diags)
	if err != nil {
		t.Fatalf(err.Error())
	}

	diags = checkValidCreds(config)
	err = checkDiags(diags)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestConfigLoadAndValidate_credsNoImpersonation(t *testing.T) {
	config := &apiClient{
		Credentials: testFakeCredentialsPath,
	}

	diags := config.loadAndValidate(context.Background())
	err := checkDiags(diags)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestConfigOauthScopes_custom(t *testing.T) {
	config := &apiClient{
		Credentials:           testFakeCredentialsPath,
		ClientScopes:          []string{"https://www.googleapis.com/auth/admin/directory"},
		ImpersonatedUserEmail: "my-fake-email@example.com",
	}

	diags := config.loadAndValidate(context.Background())
	err := checkDiags(diags)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(config.ClientScopes) != 1 {
		t.Fatalf("expected 1 scope, got %d scopes: %v", len(config.ClientScopes), config.ClientScopes)
	}
	if config.ClientScopes[0] != "https://www.googleapis.com/auth/admin/directory" {
		t.Fatalf("expected scope to be %q, got %q", "https://www.googleapis.com/auth/admin/directory", config.ClientScopes[0])
	}
}

func TestConfigLoadAndValidate_accessTokenInvalid(t *testing.T) {
	config := &apiClient{
		AccessToken:           "abcdefghijklmnopqrstuvwxyz",
		Customer:              os.Getenv("GOOGLEWORKSPACE_CUSTOMER_ID"),
		ImpersonatedUserEmail: os.Getenv("GOOGLEWORKSPACE_IMPERSONATED_USER_EMAIL"),
		ClientScopes:          []string{"https://www.googleapis.com/auth/admin.directory.domain"},
	}

	config.loadAndValidate(context.Background())
	diags := checkValidCreds(config)
	err := checkDiags(diags)
	if err == nil {
		t.Fatalf("expected error, but got nil")
	}
}

func TestConfigLoadAndValidate_accessToken(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip(fmt.Sprintf("Network access not allowed; use TF_ACC=1 to enable"))
	}

	testAccPreCheck(t)

	creds := getTestCredsFromEnv()
	gcpConfig := &apiClient{
		Credentials:  creds,
		ClientScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
	}

	diags := gcpConfig.loadAndValidate(context.Background())
	err := checkDiags(diags)
	if err != nil {
		t.Fatalf(err.Error())
	}

	iamCredsService, err := iamcredentials.NewService(context.Background(), option.WithHTTPClient(gcpConfig.client))
	if err != nil {
		t.Fatalf(err.Error())
	}
	serviceAccount := fmt.Sprintf("projects/-/serviceAccounts/%s", os.Getenv("GOOGLEWORKSPACE_IMPERSONATED_SERVICE_ACCOUNT"))
	tokenRequest := &iamcredentials.GenerateAccessTokenRequest{
		Lifetime: "300s",
		Scope:    []string{"https://www.googleapis.com/auth/cloud-platform"},
	}
	at, err := iamCredsService.Projects.ServiceAccounts.GenerateAccessToken(serviceAccount, tokenRequest).Do()
	if err != nil {
		t.Fatalf(err.Error())
	}

	config := &apiClient{
		AccessToken:           at.AccessToken,
		Customer:              os.Getenv("GOOGLEWORKSPACE_CUSTOMER_ID"),
		ImpersonatedUserEmail: os.Getenv("GOOGLEWORKSPACE_IMPERSONATED_USER_EMAIL"),
		ServiceAccount:        os.Getenv("GOOGLEWORKSPACE_IMPERSONATED_SERVICE_ACCOUNT"),
	}

	diags = config.loadAndValidate(context.Background())
	err = checkDiags(diags)
	if err != nil {
		t.Fatalf(err.Error())
	}

	diags = checkValidCreds(config)
	err = checkDiags(diags)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

// TestConfigLoadAndValidate_accessTokenOnly covers a scenario where:
// 1. A service account is given an Admin Role in Google Workspace directly (no impersonation used in this test)
// 2. That role gives it Admin API privileges to query the groups endpoint of the Admin API - `Groups Admin role`
// The provider will then only need to be configured with the customer ID and an access token for that service account
func TestConfigLoadAndValidate_accessTokenOnly(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip(fmt.Sprintf("Network access not allowed; use TF_ACC=1 to enable"))
	}

	testAccPreCheck(t)

	// Get access token for the service account
	// --- Use service account credentials to request the access token
	// --- Request the `/auth/admin.directory.group` scope as it matches privileges in the Groups Admin role
	credsFile := getTestCredsFromEnv()

	contents, _, err := pathOrContents(credsFile)
	if err != nil {
		t.Fatalf("could not get credentials: %s", err.Error())
	}

	credParams := googleoauth.CredentialsParams{
		Scopes: []string{"https://www.googleapis.com/auth/admin.directory.group"},
	}

	creds, err := googleoauth.CredentialsFromJSONWithParams(context.Background(), []byte(contents), credParams)
	if err != nil {
		t.Fatalf("could not get oauth2 credentials: %s", err.Error())
	}

	at, err := creds.TokenSource.Token()
	if err != nil {
		t.Fatalf("could not get token from oauth2 credentials: %s", err.Error())
	}

	// Configure the provider with the scoped access token from above + the customer ID
	config := &apiClient{
		AccessToken: at.AccessToken,
		Customer:    os.Getenv("GOOGLEWORKSPACE_CUSTOMER_ID"),
	}

	diags := config.loadAndValidate(context.Background())
	err = checkDiags(diags)
	if err != nil {
		t.Fatalf(err.Error())
	}

	diags = checkValidCredsGroupAdmin(config)
	err = checkDiags(diags)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func checkValidCreds(config *apiClient) diag.Diagnostics {
	var diags diag.Diagnostics

	directoryService, diags := config.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	_, err := directoryService.Customers.Get(config.Customer).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

// checkValidCredsGroupAdmin makes an arbitary API call to check the auth is set correctly.
// It makes a groups-related API call, to be used when testing auth relted to service accounts
// given the Group Admin role directly (no impersonisation done during the auth)
func checkValidCredsGroupAdmin(config *apiClient) diag.Diagnostics {
	var diags diag.Diagnostics

	directoryService, diags := config.NewDirectoryService()
	if diags.HasError() {
		return diags
	}
	groupsService := directoryService.Groups
	_, err := groupsService.List().Customer(config.Customer).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
