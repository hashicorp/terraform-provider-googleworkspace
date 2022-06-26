package googleworkspace

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceUserCustomSchemaAttributes_basic(t *testing.T) {
	t.Parallel()

	primaryEmail := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceUserCustomSchemaAttributes_basic(primaryEmail),
			},
			{
				ResourceName:      "googleworkspace_user_custom_schema_attributes.my-schema-attributes",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceUserCustomSchemaAttributes_full(t *testing.T) {
	t.Parallel()

	primaryEmail := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceUserCustomSchemaAttributes_full(primaryEmail),
			},
			{
				ResourceName:      "googleworkspace_user_custom_schema_attributes.my-schema-attributes",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceUserCustomSchemaAttributes_fullUpdate(primaryEmail),
			},
			{
				ResourceName:      "googleworkspace_user_custom_schema_attributes.my-schema-attributes",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceUserCustomSchemaAttributes_basic(primaryEmail string) string {
	return fmt.Sprintf(`
resource "googleworkspace_user_custom_schema_attributes" "my-schema-attributes" {
  primary_email = "%s"

  custom_schemas {
    schema_name = "SchemaA"
    schema_values = {
      "field" = jsonencode("value")
    }
  }
}
`, primaryEmail)
}

func testAccResourceUserCustomSchemaAttributes_full(primaryEmail string) string {
	return fmt.Sprintf(`
resource "googleworkspace_user_custom_schema_attributes" "my-schema-attributes" {
  primary_email = "%s"

  custom_schemas {
    schema_name = "SchemaA"
    schema_values = {
      "field" = jsonencode("valueA")
    }
  }

  custom_schemas {
    schema_name = "SchemaB"
    schema_values = {
      "field" = jsonencode("123")
      "attribute" = jsonencode(["valueA","valueB"])
    }
  }

  custom_schemas {
    schema_name = "SchemaC"
    schema_values = {
      "attribute" = jsonencode("valueA")
    }
  }
}
`, primaryEmail)
}

func testAccResourceUserCustomSchemaAttributes_fullUpdate(primaryEmail string) string {
	return fmt.Sprintf(`
resource "googleworkspace_user_custom_schema_attributes" "my-schema-attributes" {
  primary_email = "%s"

  custom_schemas {
    schema_name = "SchemaA"
    schema_values = {
      "field" = jsonencode("valueB")
    }
  }

  custom_schemas {
    schema_name = "SchemaB"
    schema_values = {
      "field" = jsonencode("123")
      "attribute" = jsonencode(["valueA","valueB","valueC"])
    }
  }

  custom_schemas {
    schema_name = "SchemaC"
    schema_values = {
      "attribute" = jsonencode("valueB")
    }
  }
}
`, primaryEmail)
}
