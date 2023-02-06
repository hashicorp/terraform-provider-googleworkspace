// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceGroupMember_basic(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	testGroupVals := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"groupEmail": fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccResourceGroupMemberExists("googleworkspace_group_member.my-group-member"),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGroupMember_basic(testGroupVals),
			},
			{
				ResourceName:            "googleworkspace_group_member.my-group-member",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
		},
	})
}

func TestAccResourceGroupMember_full(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	testGroupVals := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"groupEmail": fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccResourceGroupMemberExists("googleworkspace_group_member.my-group-member"),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGroupMember_full(testGroupVals),
			},
			{
				ResourceName:            "googleworkspace_group_member.my-group-member",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
			{
				Config: testAccResourceGroupMember_fullUpdate(testGroupVals),
			},
			{
				ResourceName:            "googleworkspace_group_member.my-group-member",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
		},
	})
}

func testAccResourceGroupMemberExists(resource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("%s key not found in state", resource)
		}

		client, err := googleworkspaceTestClient()
		if err != nil {
			return err
		}

		directoryService, diags := client.NewDirectoryService()
		if diags.HasError() {
			return fmt.Errorf("Error creating directory service %+v", diags)
		}

		membersService, diags := GetMembersService(directoryService)
		if diags.HasError() {
			return fmt.Errorf("Error getting group members service %+v", diags)
		}

		parts := strings.Split(rs.Primary.ID, "/")

		// id is of format "groups/<group_id>/members/<member_id>"
		if len(parts) != 4 {
			return fmt.Errorf("Group Member Id (%s) is not of the correct format (groups/<group_id>/members/<member_id>)", rs.Primary.ID)
		}

		_, err = membersService.Get(parts[1], parts[3]).Do()
		if err == nil {
			return fmt.Errorf("Group Member still exists (%s)", rs.Primary.ID)
		}

		return nil
	}
}

func testAccResourceGroupMember_basic(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{groupEmail}@%{domainName}"
}

resource "googleworkspace_user" "my-new-user" {
  primary_email = "%{userEmail}@%{domainName}"
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
`, testGroupVals)
}

func testAccResourceGroupMember_full(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{groupEmail}@%{domainName}"
}

resource "googleworkspace_user" "my-new-user" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_group_member" "my-group-member" {
  group_id = googleworkspace_group.my-group.id
  email = googleworkspace_user.my-new-user.primary_email

  role = "MANAGER"
  type = "USER"
  delivery_settings = "ALL_MAIL"
}
`, testGroupVals)
}

func testAccResourceGroupMember_fullUpdate(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{groupEmail}@%{domainName}"
}

resource "googleworkspace_user" "my-new-user" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_group_member" "my-group-member" {
  group_id = googleworkspace_group.my-group.id
  email = googleworkspace_user.my-new-user.primary_email

  role = "OWNER"
  type = "USER"
  delivery_settings = "DAILY"
}
`, testGroupVals)
}
