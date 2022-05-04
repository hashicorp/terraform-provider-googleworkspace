package googleworkspace

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceGroupSettings(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	testGroupVals := map[string]interface{}{
		"domainName": domainName,
		"email":      fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceGroupSettings(testGroupVals),
				Check: resource.ComposeTestCheckFunc(
					// Check a few of the fields
					resource.TestCheckResourceAttr(
						"data.googleworkspace_group_settings.my-group-settings", "email", Nprintf("%{email}@%{domainName}", testGroupVals)),
					resource.TestCheckResourceAttr(
						"data.googleworkspace_group_settings.my-group-settings", "name", testGroupVals["email"].(string)),
					resource.TestCheckResourceAttr(
						"data.googleworkspace_group_settings.my-group-settings", "description", ""),
					resource.TestCheckResourceAttr(
						"data.googleworkspace_group_settings.my-group-settings", "who_can_join", "CAN_REQUEST_TO_JOIN"),
					resource.TestCheckResourceAttr(
						"data.googleworkspace_group_settings.my-group-settings", "who_can_view_membership", "ALL_MEMBERS_CAN_VIEW"),
					resource.TestCheckResourceAttr(
						"data.googleworkspace_group_settings.my-group-settings", "who_can_view_group", "ALL_MEMBERS_CAN_VIEW"),
					resource.TestCheckResourceAttr(
						"data.googleworkspace_group_settings.my-group-settings", "allow_external_members", "false"),
				),
			},
		},
	})
}

func testAccDataSourceGroupSettings(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{email}@%{domainName}"
}

data "googleworkspace_group_settings" "my-group-settings" {
  email = googleworkspace_group.my-group.email
}
`, testGroupVals)
}
