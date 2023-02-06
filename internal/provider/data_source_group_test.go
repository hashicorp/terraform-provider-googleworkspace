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

func TestAccDataSourceGroup_withId(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	testGroupVals := map[string]interface{}{
		"domainName": domainName,
		"email":      fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceGroup_withId(testGroupVals),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.googleworkspace_group.my-new-group", "email", Nprintf("%{email}@%{domainName}", testGroupVals)),
					resource.TestCheckResourceAttr(
						"data.googleworkspace_group.my-new-group", "name", testGroupVals["email"].(string)),
				),
			},
		},
	})
}

func TestAccDataSourceGroup_withEmail(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	testGroupVals := map[string]interface{}{
		"domainName": domainName,
		"email":      fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceGroup_withEmail(testGroupVals),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.googleworkspace_group.my-new-group", "email", Nprintf("%{email}@%{domainName}", testGroupVals)),
					resource.TestCheckResourceAttr(
						"data.googleworkspace_group.my-new-group", "name", testGroupVals["email"].(string)),
				),
			},
		},
	})
}

func testAccDataSourceGroup_withId(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-new-group" {
  email = "%{email}@%{domainName}"
}

data "googleworkspace_group" "my-new-group" {
  id = googleworkspace_group.my-new-group.id
}
`, testGroupVals)
}

func testAccDataSourceGroup_withEmail(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-new-group" {
  email = "%{email}@%{domainName}"
}

data "googleworkspace_group" "my-new-group" {
  email = googleworkspace_group.my-new-group.email
}
`, testGroupVals)
}
