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

func TestAccDataSourceGroupMembers(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	testGroupVals := map[string]interface{}{
		"userEmail":  fmt.Sprintf("tf-test-%s@%s", acctest.RandString(10), domainName),
		"groupEmail": fmt.Sprintf("tf-test-%s@%s", acctest.RandString(10), domainName),
		"password":   acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceGroupMembers(testGroupVals),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.googleworkspace_group_members.my-group-members", "members.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.googleworkspace_group_members.my-group-members", "members.*", map[string]string{
							"email": Nprintf("%{userEmail}", testGroupVals),
							"role":  "MEMBER",
							"type":  "USER",
						}),
				),
			},
		},
	})
}

func TestAccDataSourceGroupMembersNested(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	testNestedGroupVals := map[string]interface{}{
		"userEmail":     fmt.Sprintf("tf-test-%s@%s", acctest.RandString(10), domainName),
		"subUserEmail":  fmt.Sprintf("tf-test-%s@%s", acctest.RandString(10), domainName),
		"groupEmail":    fmt.Sprintf("tf-test-%s@%s", acctest.RandString(10), domainName),
		"subGroupEmail": fmt.Sprintf("tf-test-%s@%s", acctest.RandString(10), domainName),
		"password":      acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceNestedGroupMembers(testNestedGroupVals),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.googleworkspace_group_members.nested-group-members", "members.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.googleworkspace_group_members.nested-group-members", "members.*", map[string]string{
							"email": Nprintf("%{userEmail}", testNestedGroupVals),
							"role":  "MEMBER",
							"type":  "USER",
						}),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.googleworkspace_group_members.nested-group-members", "members.*", map[string]string{
							"email": Nprintf("%{subGroupEmail}", testNestedGroupVals),
							"role":  "MEMBER",
							"type":  "GROUP",
						}),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.googleworkspace_group_members.nested-group-members", "members.*", map[string]string{
							"email": Nprintf("%{subUserEmail}", testNestedGroupVals),
							"role":  "MEMBER",
							"type":  "USER",
						}),
				),
			},
		},
	})
}

func testAccDataSourceGroupMembers(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{groupEmail}"
}

resource "googleworkspace_user" "my-new-user" {
  primary_email = "%{userEmail}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_group_member" "my-group-member" {
  group_id = googleworkspace_group.my-group.id
  email = googleworkspace_user.my-new-user.primary_email
}

data "googleworkspace_group_members" "my-group-members" {
  group_id = googleworkspace_group.my-group.id

  depends_on = [googleworkspace_group_member.my-group-member]
}
`, testGroupVals)
}

func testAccDataSourceNestedGroupMembers(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "parent-group" {
  email = "%{groupEmail}"
}

resource "googleworkspace_group" "sub-group" {
  email = "%{subGroupEmail}"
}
resource "googleworkspace_user" "user" {
  primary_email = "%{userEmail}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_user" "sub-user" {
  primary_email = "%{subUserEmail}"
  password = "%{password}"

  name {
    family_name = "Schrute"
    given_name = "Dwight"
  }
}

resource "googleworkspace_group_member" "user-member" {
  group_id = googleworkspace_group.parent-group.id
  email = googleworkspace_user.user.primary_email
}

resource "googleworkspace_group_member" "user-sub-member" {
  group_id = googleworkspace_group.sub-group.id
  email = googleworkspace_user.sub-user.primary_email
}

resource "googleworkspace_group_member" "sub-group-member" {
  group_id = googleworkspace_group.parent-group.id
  email = googleworkspace_group.sub-group.email
  type = "GROUP"
}

data "googleworkspace_group_members" "nested-group-members" {
  group_id = googleworkspace_group.parent-group.id
  include_derived_membership = true
  depends_on = [
    googleworkspace_group_member.user-member,
    googleworkspace_group_member.user-sub-member,
    googleworkspace_group_member.sub-group-member,
  ]
}
`, testGroupVals)
}
