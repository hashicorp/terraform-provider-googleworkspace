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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRole_basic(fmt.Sprintf("tf-test-%s", acctest.RandString(10)), "test"),
			},
			{
				ResourceName:      "googleworkspace_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceRole_full(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRole_basic(fmt.Sprintf("tf-test-%s", acctest.RandString(10)), "test"),
			},
			{
				ResourceName:      "googleworkspace_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRole_update(fmt.Sprintf("tf-test-%s", acctest.RandString(10)), "update"),
			},
			{
				ResourceName:      "googleworkspace_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRole_basic(name, description string) string {
	return fmt.Sprintf(`
resource "googleworkspace_role" "test" {
  name = "%s"
  description = "%s"
 
  privileges {
    service_id = "02w5ecyt3pkeyqi"
    privilege_name = "MANAGE_PLAY_FOR_WORK_STORE"
  }

  privileges {
    service_id = "02w5ecyt3pkeyqi"
    privilege_name = "MANAGE_ENTERPRISE_PRIVATE_APPS"
  }
}
`, name, description)
}

func testAccRole_update(name, description string) string {
	return fmt.Sprintf(`
resource "googleworkspace_role" "test" {
  name = "%s"
  description = "%s"
 
  privileges {
    service_id = "02w5ecyt3pkeyqi"
    privilege_name = "MANAGE_PLAY_FOR_WORK_STORE"
  }

  privileges {
    service_id = "02w5ecyt3pkeyqi"
    privilege_name = "MANAGE_ENTERPRISE_PRIVATE_APPS"
  }

  privileges {
    service_id = "02w5ecyt3pkeyqi"
    privilege_name = "MANAGE_EXTERNALLY_HOSTED_APK_UPLOAD_IN_PLAY"
  }
}
`, name, description)
}
