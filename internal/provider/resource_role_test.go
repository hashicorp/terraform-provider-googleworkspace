package googleworkspace

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceRole_basic(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRole_basic(fmt.Sprintf("tf-test-%s", acctest.RandString(10)), "test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googleworkspace_role.test", "privileges.#", "9"),
				),
			},
			{
				ResourceName:            "googleworkspace_role.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
		},
	})
}

func TestAccResourceRole_full(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRole_basic(fmt.Sprintf("tf-test-%s", acctest.RandString(10)), "test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googleworkspace_role.test", "privileges.#", "9"),
				),
			},
			{
				ResourceName:            "googleworkspace_role.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
			{
				Config: testAccRole_update(fmt.Sprintf("tf-test-%s", acctest.RandString(10)), "update"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googleworkspace_role.test", "privileges.#", "13"),
				),
			},
			{
				ResourceName:            "googleworkspace_role.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
		},
	})
}

func testAccRole_basic(name, description string) string {
	return fmt.Sprintf(`
data "googleworkspace_privileges" "privileges" {}

locals {
  read_only_privileges = [
    for priv in data.googleworkspace_privileges.privileges.items : priv
    if length(regexall("READ", priv.privilege_name)) > 0
  ]
  privileges_by_service_name = [
    for priv in data.googleworkspace_privileges.privileges.items : priv
    if priv.service_name == "gmail"
  ]
}

resource "googleworkspace_role" "test" {
  name = "%s"
  description = "%s"
 
  dynamic "privileges" {
    for_each = local.read_only_privileges
    content {
      service_id = privileges.value["service_id"]
      privilege_name = privileges.value["privilege_name"]
    }
  }
}
`, name, description)
}

func testAccRole_update(name, description string) string {
	return fmt.Sprintf(`
data "googleworkspace_privileges" "privileges" {}

locals {
  read_only_privileges = [
    for priv in data.googleworkspace_privileges.privileges.items : priv
    if length(regexall("READ", priv.privilege_name)) > 0
  ]
  privileges_by_service_name = [
    for priv in data.googleworkspace_privileges.privileges.items : priv
    if priv.service_name == "gmail"
  ]
}

resource "googleworkspace_role" "test" {
  name = "%s"
  description = "%s"

  dynamic "privileges" {
    for_each = local.read_only_privileges
    content {
      service_id = privileges.value["service_id"]
      privilege_name = privileges.value["privilege_name"]
    }
  }

  dynamic "privileges" {
    for_each = local.privileges_by_service_name
    content {
      service_id = privileges.value["service_id"]
      privilege_name = privileges.value["privilege_name"]
    }
  }
}
`, name, description)
}
