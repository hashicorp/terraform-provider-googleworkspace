package googleworkspace

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceRoleAssignment(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRoleAssignment(""),
			},
			{
				ResourceName:      "googleworkspace_role_assignment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceRoleAssignment(roleId string) string {
	return fmt.Sprintf(`
resource "googleworkspace_role_assignment" "test" {
  role_id = "%s"
}
`, roleId)
}
