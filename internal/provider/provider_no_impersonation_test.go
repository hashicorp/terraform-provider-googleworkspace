// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// this test requires the service account in the credentials file to have
// GROUPS ADMIN role assignment. Use the rest api
// "Try this API" https://developers.google.com/admin-sdk/directory/reference/rest/v1/roleAssignments/insert
// you will need to determine the roleId of GROUPS ADMIN and use the client_id
// in the credentials file as 'assignedTo', 'scopeType' should be 'CUSTOMER'
func TestAccResourceGroup_noImpersonation(t *testing.T) {
	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	impersonation := getTestImpersonatedUserFromEnv()
	os.Unsetenv("GOOGLEWORKSPACE_IMPERSONATED_USER_EMAIL")
	t.Cleanup(func() {
		os.Setenv("GOOGLEWORKSPACE_IMPERSONATED_USER_EMAIL", impersonation)
	})

	testGroupVals := map[string]interface{}{
		"domainName": domainName,
		"email":      fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGroup_basic(testGroupVals),
			},
			{
				ResourceName:            "googleworkspace_group.my-group",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
		},
	})
}

// this test requires the service account in the credentials file to have
// USER MANAGEMENT ADMIN role assignment. Use the rest api
// "Try this API" https://developers.google.com/admin-sdk/directory/reference/rest/v1/roleAssignments/insert
// you will need to determine the roleId of _USER_MANAGEMENT_ADMIN_ROLE and use the client_id
// in the credentials file as 'assignedTo', 'scopeType' should be 'CUSTOMER'
func TestAccResourceUser_noImpersonation(t *testing.T) {
	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	impersonation := getTestImpersonatedUserFromEnv()
	os.Unsetenv("GOOGLEWORKSPACE_IMPERSONATED_USER_EMAIL")
	t.Cleanup(func() {
		os.Setenv("GOOGLEWORKSPACE_IMPERSONATED_USER_EMAIL", impersonation)
	})

	testUserVals := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceUser_basic(testUserVals),
			},
			{
				ResourceName:            "googleworkspace_user.my-new-user",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag", "password"},
			},
		},
	})
}
