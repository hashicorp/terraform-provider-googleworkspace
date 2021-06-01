package googleworkspace

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceRole(t *testing.T) {
	t.Parallel()

	name := "cloudjobdiscovery.admin"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRole(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.googleworkspace_role.test", "name", "nope"),
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
