package googleworkspace

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceRole_full(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRole_basic(map[string]interface{}{
					"name":        fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
					"description": "test",
				}),
			},
			{
				ResourceName:      "googleworkspace_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRole_update(map[string]interface{}{
					"name":        fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
					"description": "test",
				}),
			},
			{
				ResourceName:      "googleworkspace_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRole_basic(data map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_role" "test" {
  name = "%{name}"
  description = "%{description}"
 
  privileges {
    service_id = "02w5ecyt3pkeyqi"
    name = "MANAGE_PLAY_FOR_WORK_STORE"
  }

  privileges {
	service_id = "02w5ecyt3pkeyqi"
    name = "MANAGE_ENTERPRISE_PRIVATE_APPS"
  }
}
`, data)
}

func testAccRole_update(data map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_role" "test" {
  name = "%{name}"
  description = "%{description}"
 
  privileges {
    service_id = "02w5ecyt3pkeyqi"
    name = "MANAGE_PLAY_FOR_WORK_STORE"
  }

  privileges {
	service_id = "02w5ecyt3pkeyqi"
    name = "MANAGE_ENTERPRISE_PRIVATE_APPS"
  }

  privileges {
	service_id = "02w5ecyt3pkeyqi"
	name = "MANAGE_EXTERNALLY_HOSTED_APK_UPLOAD_IN_PLAY"
  }
}
`, data)
}
