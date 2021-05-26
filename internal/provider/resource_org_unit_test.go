package googleworkspace

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceOrgUnit_basic(t *testing.T) {
	t.Parallel()

	ouName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccResourceOrgUnitMemberExists("googleworkspace_org_unit.my-org-unit"),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceOrgUnit_basic(ouName),
			},
			{
				ResourceName:      "googleworkspace_org_unit.my-org-unit",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceOrgUnit_full(t *testing.T) {
	t.Parallel()

	ouName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccResourceOrgUnitMemberExists("googleworkspace_org_unit.my-org-unit"),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceOrgUnit_full(ouName),
			},
			{
				ResourceName:      "googleworkspace_org_unit.my-org-unit",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceOrgUnit_fullUpdate(ouName),
			},
			{
				ResourceName:      "googleworkspace_org_unit.my-org-unit",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceOrgUnitMemberExists(resource string) resource.TestCheckFunc {
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

		orgUnitsService, diags := GetOrgUnitsService(directoryService)
		if diags.HasError() {
			return fmt.Errorf("Error getting org units service %+v", diags)
		}

		_, err = orgUnitsService.Get(client.Customer, rs.Primary.ID).Do()
		if err == nil {
			return fmt.Errorf("Org Unit still exists (%s)", rs.Primary.ID)
		}

		return nil
	}
}

func testAccResourceOrgUnit_basic(ouName string) string {
	return fmt.Sprintf(`
resource "googleworkspace_org_unit" "my-org-unit" {
  name = "%s"
  parent_org_unit_path = "/"
}
`, ouName)
}

func testAccResourceOrgUnit_full(ouName string) string {
	return fmt.Sprintf(`
resource "googleworkspace_org_unit" "my-org-unit" {
  name = "%s"
  description = "my test description"
  block_inheritance = true
  parent_org_unit_path = "/"
}
`, ouName)
}

func testAccResourceOrgUnit_fullUpdate(ouName string) string {
	return fmt.Sprintf(`
resource "googleworkspace_org_unit" "my-org-unit" {
  name = "%s-new"
  description = "my new test description"
  block_inheritance = false
  parent_org_unit_path = "/"
}
`, ouName)
}
