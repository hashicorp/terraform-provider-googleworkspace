package googleworkspace

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// List of prefixes used for test resource names
var testResourcePrefixes = []string{
	"tf-test",
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// sharedClientForRegion returns a common config setup needed for the sweeper
// functions for a given region
func sharedClientForRegion() (*apiClient, error) {
	creds := getTestCredsFromEnv()
	if creds == "" {
		return nil, fmt.Errorf("set credentials using any of these env variables %v", credsEnvVars)
	}

	customerId := getTestCustomerFromEnv()
	if customerId == "" {
		return nil, fmt.Errorf("set customer id with GOOGLEWORKSPACE_CUSTOMER_ID")
	}

	impersonatedUser := getTestImpersonatedUserFromEnv()
	if impersonatedUser == "" {
		return nil, fmt.Errorf("set customer id with GOOGLEWORKSPACE_IMPERSONATED_USER_EMAIL")
	}

	client := &apiClient{
		Credentials:           creds,
		Customer:              customerId,
		ImpersonatedUserEmail: impersonatedUser,
	}

	return client, nil
}

func isSweepableTestResource(resourceName string) bool {
	for _, p := range testResourcePrefixes {
		if strings.HasPrefix(resourceName, p) {
			return true
		}
	}
	return false
}
