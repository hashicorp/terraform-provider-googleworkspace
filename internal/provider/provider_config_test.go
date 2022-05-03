package googleworkspace

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"google.golang.org/api/iamcredentials/v1"
	"google.golang.org/api/option"
)

var testOauthScopesDirectory = types.List{
	Elems: []attr.Value{
		types.String{Value: "https://www.googleapis.com/auth/admin.directory.customer"},
	},
	ElemType: types.StringType,
}

func TestConfigLoadAndValidate_credsInvalidJSON(t *testing.T) {
	diags := diag.Diagnostics{}
	config := &providerData{
		Credentials:           types.String{Value: "{this is not json}"},
		ImpersonatedUserEmail: types.String{Value: "my-fake-email@example.com"},
		OauthScopes:           testOauthScopesDirectory,
	}

	_ = authenticateClient(context.Background(), *config, &diags)
	if !diags.HasError() {
		t.Fatalf("expected error, but got nil")
	}
}

func TestConfigLoadAndValidate_credsJSON(t *testing.T) {
	diags := diag.Diagnostics{}
	contents, err := ioutil.ReadFile(testFakeCredentialsPath)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	config := &providerData{
		Credentials:           types.String{Value: string(contents)},
		ImpersonatedUserEmail: types.String{Value: "my-fake-email@example.com"},
		OauthScopes:           testOauthScopesDirectory,
	}

	_ = authenticateClient(context.Background(), *config, &diags)
	if diags.HasError() {
		t.Fatalf(getDiagErrors(diags))
	}
}

func TestConfigLoadAndValidate_credsFromFile(t *testing.T) {
	diags := diag.Diagnostics{}
	config := &providerData{
		Credentials:           types.String{Value: testFakeCredentialsPath},
		ImpersonatedUserEmail: types.String{Value: "my-fake-email@example.com"},
		OauthScopes:           testOauthScopesDirectory,
	}

	_ = authenticateClient(context.Background(), *config, &diags)
	if diags.HasError() {
		t.Fatalf(getDiagErrors(diags))
	}
}

func TestAccConfigLoadAndValidate_credsFromEnv(t *testing.T) {
	diags := diag.Diagnostics{}
	if os.Getenv("TF_ACC") == "" {
		t.Skip(fmt.Sprintf("Network access not allowed; use TF_ACC=1 to enable"))
	}

	testAccPreCheck(t)

	creds := getTestCredsFromEnv()
	config := &providerData{
		Credentials:           types.String{Value: creds},
		Customer:              types.String{Value: os.Getenv("GOOGLEWORKSPACE_CUSTOMER_ID")},
		ImpersonatedUserEmail: types.String{Value: os.Getenv("GOOGLEWORKSPACE_IMPERSONATED_USER_EMAIL")},
		OauthScopes:           testOauthScopesDirectory,
	}

	testClient := authenticateClient(context.Background(), *config, &diags)
	if diags.HasError() {
		t.Fatalf(getDiagErrors(diags))
	}

	testProv := provider{
		client:   testClient,
		customer: os.Getenv("GOOGLEWORKSPACE_CUSTOMER_ID"),
	}

	checkValidCreds(&testProv, &diags)
	if diags.HasError() {
		t.Fatalf(getDiagErrors(diags))
	}
}

func TestConfigLoadAndValidate_credsNoImpersonation(t *testing.T) {
	diags := diag.Diagnostics{}
	config := &providerData{
		Credentials: types.String{Value: testFakeCredentialsPath},
		OauthScopes: testOauthScopesDirectory,
	}

	_ = authenticateClient(context.Background(), *config, &diags)
	if diags.HasError() {
		t.Fatalf(getDiagErrors(diags))
	}
}

func TestConfigOauthScopes_custom(t *testing.T) {
	diags := diag.Diagnostics{}
	config := &providerData{
		Credentials:           types.String{Value: testFakeCredentialsPath},
		ImpersonatedUserEmail: types.String{Value: "my-fake-email@example.com"},
		OauthScopes:           testOauthScopesDirectory,
	}

	_ = authenticateClient(context.Background(), *config, &diags)
	if diags.HasError() {
		t.Fatalf(getDiagErrors(diags))
	}

	if len(config.OauthScopes.Elems) != 1 {
		t.Fatalf("expected 1 scope, got %d scopes: %v", len(config.OauthScopes.Elems), config.OauthScopes)
	}
	if config.OauthScopes.Elems[0].(types.String).Value != "https://www.googleapis.com/auth/admin/directory" {
		t.Fatalf("expected scope to be %s, got %s", "https://www.googleapis.com/auth/admin/directory", config.OauthScopes.Elems[0].(types.String).Value)
	}
}

func TestConfigLoadAndValidate_accessTokenInvalid(t *testing.T) {
	diags := diag.Diagnostics{}
	config := &providerData{
		Credentials:           types.String{Value: "abcdefghijklmnopqrstuvwxyz"},
		Customer:              types.String{Value: os.Getenv("GOOGLEWORKSPACE_CUSTOMER_ID")},
		ImpersonatedUserEmail: types.String{Value: os.Getenv("GOOGLEWORKSPACE_IMPERSONATED_USER_EMAIL")},
		OauthScopes: types.List{
			Elems: []attr.Value{
				types.String{Value: "https://www.googleapis.com/auth/admin.directory.domain"},
			},
			ElemType: types.StringType,
		},
	}

	testClient := authenticateClient(context.Background(), *config, &diags)
	testProv := &provider{
		client:   testClient,
		customer: os.Getenv("GOOGLEWORKSPACE_CUSTOMER_ID"),
	}

	checkValidCreds(testProv, &diags)
	if !diags.HasError() {
		t.Fatalf("expected error, but got nil")
	}
}

func TestConfigLoadAndValidate_accessToken(t *testing.T) {
	diags := diag.Diagnostics{}
	if os.Getenv("TF_ACC") == "" {
		t.Skip(fmt.Sprintf("Network access not allowed; use TF_ACC=1 to enable"))
	}

	testAccPreCheck(t)

	creds := getTestCredsFromEnv()
	gcpConfig := &providerData{
		Credentials: types.String{Value: creds},
		OauthScopes: types.List{
			Elems: []attr.Value{
				types.String{Value: "https://www.googleapis.com/auth/cloud-platform"},
			},
		},
	}

	gcpClient := authenticateClient(context.Background(), *gcpConfig, &diags)
	if diags.HasError() {
		t.Fatalf(getDiagErrors(diags))
	}

	iamCredsService, err := iamcredentials.NewService(context.Background(), option.WithHTTPClient(gcpClient))
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

	gwConfig := &providerData{
		AccessToken:           types.String{Value: at.AccessToken},
		Customer:              types.String{Value: os.Getenv("GOOGLEWORKSPACE_CUSTOMER_ID")},
		ImpersonatedUserEmail: types.String{Value: os.Getenv("GOOGLEWORKSPACE_IMPERSONATED_USER_EMAIL")},
		ServiceAccount:        types.String{Value: os.Getenv("GOOGLEWORKSPACE_IMPERSONATED_SERVICE_ACCOUNT")},
	}

	gwClient := authenticateClient(context.Background(), *gwConfig, &diags)
	if diags.HasError() {
		t.Fatalf(err.Error())
	}

	testProv := &provider{
		client:   gwClient,
		customer: os.Getenv("GOOGLEWORKSPACE_CUSTOMER_ID"),
	}

	checkValidCreds(testProv, &diags)
	if diags.HasError() {
		t.Fatalf(getDiagErrors(diags))
	}
}

func checkValidCreds(prov *provider, diags *diag.Diagnostics) {
	directoryService := prov.NewDirectoryService(diags)
	if diags.HasError() {
		return
	}

	_, err := directoryService.Customers.Get(prov.customer).Do()
	if err != nil {
		diags.AddError("get customer failed", err.Error())
		return
	}

	return
}
