package googleworkspace

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceSchema_withId(t *testing.T) {

	schemaName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSchema_withId(schemaName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.googleworkspace_schema.my-schema", "schema_name", schemaName),
					resource.TestCheckResourceAttr(
						"data.googleworkspace_schema.my-schema", "fields.0.field_name", "birthday"),
					resource.TestCheckResourceAttr(
						"data.googleworkspace_schema.my-schema", "fields.0.field_type", "DATE"),
				),
			},
		},
	})
}

func TestAccDataSourceSchema_withName(t *testing.T) {

	schemaName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSchema_withName(schemaName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.googleworkspace_schema.my-schema", "schema_name", schemaName),
					resource.TestCheckResourceAttr(
						"data.googleworkspace_schema.my-schema", "fields.0.field_name", "birthday"),
					resource.TestCheckResourceAttr(
						"data.googleworkspace_schema.my-schema", "fields.0.field_type", "DATE"),
				),
			},
		},
	})
}

func testAccDataSourceSchema_withId(schemaName string) string {
	return fmt.Sprintf(`
resource "googleworkspace_schema" "my-schema" {
  schema_name = "%s"

  fields {
    field_name = "birthday"
    field_type = "DATE"
  }
}

data "googleworkspace_schema" "my-schema" {
  schema_name = googleworkspace_schema.my-schema.schema_id
}
`, schemaName)
}

func testAccDataSourceSchema_withName(schemaName string) string {
	return fmt.Sprintf(`
resource "googleworkspace_schema" "my-schema" {
  schema_name = "%s"

  fields {
    field_name = "birthday"
    field_type = "DATE"
  }
}

data "googleworkspace_schema" "my-schema" {
  schema_name = googleworkspace_schema.my-schema.schema_name
}
`, schemaName)
}
