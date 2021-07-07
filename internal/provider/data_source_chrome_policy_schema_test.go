package googleworkspace

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceChromePolicySchema(t *testing.T) {
	t.Parallel()

	schemaName := "chrome.printers.AllowForUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceChromePolicySchema(schemaName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.googleworkspace_chrome_policy_schema.test", "schema_name", schemaName),
					resource.TestCheckResourceAttr("data.googleworkspace_chrome_policy_schema.test", "policy_description", "Allows a printer for users in a given organization."),
					resource.TestCheckResourceAttr("data.googleworkspace_chrome_policy_schema.test", "additional_target_key_names.#", "1"),
					resource.TestCheckResourceAttr("data.googleworkspace_chrome_policy_schema.test", "additional_target_key_names.0.key", "printer_id"),
				),
			},
		},
	})
}

func testAccDataSourceChromePolicySchema(schemaName string) string {
	return fmt.Sprintf(`
data "googleworkspace_chrome_policy_schema" "test" {
  schema_name = "%s"
}
`, schemaName)
}
