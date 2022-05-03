package googleworkspace

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"log"
	"os"
	"strings"
	"testing"
)

const testFakeCredentialsPath = "./test-data/fake-creds.json"

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"googleworkspace": func() (tfprotov6.ProviderServer, error) {
		return tfsdk.NewProtocol6Server(New("test")()), nil
	},
}

var credsEnvVars = []string{
	"GOOGLEWORKSPACE_CREDENTIALS",
	"GOOGLEWORKSPACE_CLOUD_KEYFILE_JSON",
	"GOOGLEWORKSPACE_USE_DEFAULT_CREDENTIALS",
	"GOOGLE_CREDENTIALS",
}

func getTestCredsFromEnv() string {
	// Return empty string if GOOGLEWORKSPACE_USE_DEFAULT_CREDENTIALS is set to true.
	if os.Getenv("GOOGLEWORKSPACE_USE_DEFAULT_CREDENTIALS") == "true" {
		return ""
	}
	return multiEnvSearch(credsEnvVars)
}

func getTestCustomerFromEnv() string {
	return os.Getenv("GOOGLEWORKSPACE_CUSTOMER_ID")
}

func getTestImpersonatedUserFromEnv() string {
	return os.Getenv("GOOGLEWORKSPACE_IMPERSONATED_USER_EMAIL")
}

// googleworkspaceTestClient returns a common client
func googleworkspaceTestClient(ctx context.Context, diags diag.Diagnostics) *provider {
	creds := getTestCredsFromEnv()
	if creds == "" {
		diags.AddError("credentials are required", fmt.Sprintf("set credentials using any of these env variables %v", credsEnvVars))
	}

	customerId := getTestCustomerFromEnv()
	if customerId == "" {
		diags.AddError("customer id is required", "set customer id with GOOGLEWORKSPACE_CUSTOMER_ID")
	}

	impersonatedUser := getTestImpersonatedUserFromEnv()
	if impersonatedUser == "" {
		diags.AddError("impersonated user is required", "set customer id with GOOGLEWORKSPACE_IMPERSONATED_USER_EMAIL")
	}
	if diags.HasError() {
		log.Printf("[INFO][SWEEPER_LOG] error reading env variables for test provider:\n%s", getDiagErrors(diags))
		return nil
	}

	pd := &providerData{
		Credentials:           types.String{Value: creds},
		Customer:              types.String{Value: customerId},
		ImpersonatedUserEmail: types.String{Value: impersonatedUser},
		OauthScopes:           primitiveListToTypeList(types.StringType, DefaultClientScopes),
	}

	p := &provider{
		version:  "test",
		client:   authenticateClient(ctx, *pd, &diags),
		customer: customerId,
	}
	if diags.HasError() {
		log.Printf("[INFO][SWEEPER_LOG]\nerror creating test provider: %s", getDiagErrors(diags))
		return nil
	}

	return p
}

func TestProvider(t *testing.T) {
	provider := New("dev")()
	if provider == nil {
		t.Fatalf("provider is nil")
	}
}

// testAccPreCheck ensures at least one of the credentials env variables is set.
func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.

	if v := multiEnvSearch(credsEnvVars); v == "" {
		t.Fatalf("One of %s must be set for acceptance tests", strings.Join(credsEnvVars, ", "))
	}
}

func multiEnvSearch(ks []string) string {
	for _, k := range ks {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}
