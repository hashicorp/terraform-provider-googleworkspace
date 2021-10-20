package googleworkspace

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceSchema_basic(t *testing.T) {
	t.Parallel()

	schemaName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSchema_basic(schemaName),
			},
			{
				ResourceName:            "googleworkspace_schema.my-schema",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
		},
	})
}

func TestAccResourceSchema_full(t *testing.T) {
	t.Parallel()

	schemaName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSchema_full(schemaName),
			},
			{
				ResourceName:            "googleworkspace_schema.my-schema",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
			{
				Config: testAccResourceSchema_fullUpdate(schemaName),
			},
			{
				ResourceName:            "googleworkspace_schema.my-schema",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
		},
	})
}

func testAccResourceSchema_basic(schemaName string) string {
	return fmt.Sprintf(`
resource "googleworkspace_schema" "my-schema" {
  schema_name = "%s"

  fields {
    field_name = "birthday"
    field_type = "DATE"
  }
}
`, schemaName)
}

func testAccResourceSchema_full(schemaName string) string {
	return fmt.Sprintf(`
resource "googleworkspace_schema" "my-schema" {
  schema_name = "%s-updated"
  display_name = "schema test full"

  fields {
    field_name = "birthday"
    field_type = "DATE"
    read_access_type = "ADMINS_AND_SELF"
  }

  fields {
    field_name = "favorite_numbers"
    field_type = "INT64"
    multi_valued = true

    numeric_indexing_spec {
      min_value = 1.0
      max_value = 10.5
    }
  }

  fields {
    field_name = "indexed"
    field_type = "DOUBLE"
    indexed = true
  }
}
`, schemaName)
}

func testAccResourceSchema_fullUpdate(schemaName string) string {
	return fmt.Sprintf(`
resource "googleworkspace_schema" "my-schema" {
  schema_name = "%s"
  display_name = "schema test full update"

  fields {
    field_name = "birthday"
    field_type = "DATE"
  }

  fields {
    field_name = "favorite_number"
    field_type = "INT64"
    multi_valued = true

    numeric_indexing_spec {
      min_value = 1.0
      max_value = 13.2
    }
  }

  fields {
    field_name = "indexed"
    field_type = "DOUBLE"
    multi_valued = true
    indexed = true
    read_access_type = "ADMINS_AND_SELF"
  }
}
`, schemaName)
}
