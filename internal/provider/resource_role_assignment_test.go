package googleworkspace

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceRoleAssignment_basic(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	data := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleAssignment_basic(data),
			},
			{
				ResourceName:            "googleworkspace_role_assignment.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
		},
	})
}

func TestAccResourceRoleAssignment_orgUnit_invalid(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	data := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"roleName":   fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRoleAssignment_orgUnit_invalid(data),
				ExpectError: regexp.MustCompile("Attribute cannot be empty"),
			},
		},
	})
}

func TestAccResourceRoleAssignment_orgUnit(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	data := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"roleName":   fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"ouName":     fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleAssignment_orgUnit(data),
			},
			{
				ResourceName:            "googleworkspace_role_assignment.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
		},
	})
}

func testAccRoleAssignment_basic(data map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_user" "test" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name = {
    family_name = "Scott"
    given_name = "Michael"
  }
}

data "googleworkspace_role" "test" {
  name = "_GROUPS_ADMIN_ROLE"
}

resource "googleworkspace_role_assignment" "test" {
  role_id = data.googleworkspace_role.test.id
  assigned_to = googleworkspace_user.test.id
}
`, data)
}

func testAccRoleAssignment_orgUnit_invalid(data map[string]interface{}) string {
	return Nprintf(`
data "googleworkspace_privileges" "privileges" {}

locals {
  org_scopable_privileges = [
    for priv in data.googleworkspace_privileges.privileges.items : priv
    if priv.is_org_unit_scopable
  ]
}

resource "googleworkspace_role" "test" {
  name = "%{roleName}"
 
  dynamic "privileges" {
    for_each = local.org_scopable_privileges
    content {
      service_id = privileges.value["service_id"]
      privilege_name = privileges.value["privilege_name"]
    }
  }
}

resource "googleworkspace_user" "test" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name = {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_role_assignment" "test" {
  role_id = googleworkspace_role.test.id
  assigned_to = googleworkspace_user.test.id
  scope_type = "ORG_UNIT"
}
`, data)
}

func testAccRoleAssignment_orgUnit(data map[string]interface{}) string {
	return Nprintf(`
data "googleworkspace_privileges" "privileges" {}

locals {
  org_scopable_privileges = [
    for priv in data.googleworkspace_privileges.privileges.items : priv
    if priv.is_org_unit_scopable
  ]
}

resource "googleworkspace_role" "test" {
  name = "%{roleName}"
 
  dynamic "privileges" {
    for_each = local.org_scopable_privileges
    content {
      service_id = privileges.value["service_id"]
      privilege_name = privileges.value["privilege_name"]
    }
  }
}

resource "googleworkspace_user" "test" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name = {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_org_unit" "test" {
  name = "%{ouName}"
  parent_org_unit_path = "/"
}

resource "googleworkspace_role_assignment" "test" {
  role_id = googleworkspace_role.test.id
  assigned_to = googleworkspace_user.test.id
  scope_type = "ORG_UNIT"
  org_unit_id = googleworkspace_org_unit.test.id
}
`, data)
}
