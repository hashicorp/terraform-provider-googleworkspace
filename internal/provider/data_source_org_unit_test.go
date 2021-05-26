package googleworkspace

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceOrgUnit_withOrgUnitId(t *testing.T) {
	t.Parallel()

	ouName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceOrgUnit_withOrgUnitId(ouName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.googleworkspace_org_unit.my-new-org-unit", "name", ouName),
				),
			},
		},
	})
}

func TestAccDataSourceOrgUnit_withOrgUnitPath(t *testing.T) {
	t.Parallel()

	ouName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceOrgUnit_withOrgUnitPath(ouName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.googleworkspace_org_unit.my-new-org-unit", "name", ouName),
				),
			},
		},
	})
}

func testAccDataSourceOrgUnit_withOrgUnitId(ouName string) string {
	return fmt.Sprintf(`
resource "googleworkspace_org_unit" "my-new-org-unit" {
  name = "%s"
  parent_org_unit_path = "/"
}

data "googleworkspace_org_unit" "my-new-org-unit" {
  org_unit_id = googleworkspace_org_unit.my-new-org-unit.id
}
`, ouName)
}

func testAccDataSourceOrgUnit_withOrgUnitPath(ouName string) string {
	return fmt.Sprintf(`
resource "googleworkspace_org_unit" "my-new-org-unit" {
  name = "%s"
  parent_org_unit_path = "/"
}

data "googleworkspace_org_unit" "my-new-org-unit" {
  org_unit_path = googleworkspace_org_unit.my-new-org-unit.org_unit_path
}
`, ouName)
}
