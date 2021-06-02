package googleworkspace

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceRoleAssignment_customer(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	data := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
		"role":       "_GROUPS_ADMIN_ROLE",
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleAssignment_customer(data),
			},
			{
				ResourceName:      "googleworkspace_role_assignment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRoleAssignment_customer(data map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_user" "test" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name {
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
  scope_type = "CUSTOMER"
}
`, data)
}
