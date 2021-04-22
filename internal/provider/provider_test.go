package googleworkspace

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const testFakeCredentialsPath = "./test-data/fake-creds.json"

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var providerFactories = map[string]func() (*schema.Provider, error){
	"googleworkspace": func() (*schema.Provider, error) {
		return New("dev")(), nil
	},
}

var credsEnvVars = []string{
	"GOOGLEWORKSPACE_CREDENTIALS",
	"GOOGLEWORKSPACE_CLOUD_KEYFILE_JSON",
	"GOOGLEWORKSPACE_USE_DEFAULT_CREDENTIALS",
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

func TestProvider(t *testing.T) {
	if err := New("dev")().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
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

// checkDiags will check to see if any of the diags have type error, and if so, they will return the first
// error. It will print warnings until it hits an error.
func checkDiags(diags diag.Diagnostics) error {
	if diags.HasError() {
		for _, d := range diags {
			if d.Severity == diag.Error {
				return fmt.Errorf("Error: %s (%s)", d.Summary, d.Detail)
			}

			fmt.Println("Warning: ", d.Detail)
		}
	}

	return nil
}
