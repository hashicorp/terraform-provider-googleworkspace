package googleworkspace

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceRole(t *testing.T) {
	t.Parallel()

	name := "_GROUPS_ADMIN_ROLE"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRole(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.googleworkspace_role.test", "id", "11891411796361218"),
					resource.TestCheckResourceAttr("data.googleworkspace_role.test", "name", name),
					resource.TestCheckResourceAttr("data.googleworkspace_role.test", "is_system_role", "true"),
					resource.TestCheckResourceAttr("data.googleworkspace_role.test", "privileges.#", "6"),
					resource.TestCheckResourceAttr("data.googleworkspace_role.test", "privileges.0.name", "CHANGE_USER_GROUP_MEMBERSHIP"),
				),
			},
		},
	})
}

func testAccDataSourceRole(name string) string {
	return fmt.Sprintf(`
data "googleworkspace_role" "test" {
  name = "%s"
}
`, name)
}
