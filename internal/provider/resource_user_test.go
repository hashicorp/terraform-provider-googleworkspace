package googleworkspace

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceUser_basic(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	testUserVals := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceUser_basic(testUserVals),
			},
			{
				ResourceName:            "googleworkspace_user.my-new-user",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccResourceUser_noPassword(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	testUserVals := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceUser_noPassword(testUserVals),
				ExpectError: regexp.MustCompile("Password is required"),
			},
		},
	})
}

func TestAccResourceUser_full(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	testUserVals := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceUser_full(testUserVals),
			},
			{
				ResourceName:            "googleworkspace_user.my-new-user",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "hash_function"},
			},
			{
				Config: testAccResourceUser_fullUpdate(testUserVals),
			},
			{
				ResourceName:            "googleworkspace_user.my-new-user",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "hash_function"},
			},
		},
	})
}

func TestAccResourceUser_isAdmin(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	testUserVals := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceUser_isAdmin(testUserVals, "true"),
			},
			{
				ResourceName:            "googleworkspace_user.my-new-user",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: testAccResourceUser_isAdmin(testUserVals, "false"),
			},
			{
				ResourceName:            "googleworkspace_user.my-new-user",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

// this will test suspending a user, then archiving
func TestAccResourceUser_gone(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	testUserVals := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceUser_basic(testUserVals),
			},
			{
				ResourceName:            "googleworkspace_user.my-new-user",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: testAccResourceUser_suspended(testUserVals),
			},
			{
				ResourceName:            "googleworkspace_user.my-new-user",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			// We need to buy Archived User Licenses in order to test this
			//{
			//	Config: testAccResourceUser_archived(testUserVals),
			//},
			//{
			//	ResourceName:            "googleworkspace_user.my-new-user",
			//	ImportState:             true,
			//	ImportStateVerify:       true,
			//	ImportStateVerifyIgnore: []string{"password"},
			//},
		},
	})
}

func TestAccResourceUser_customSchemasAllTypes(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	testUserVals := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceUser_customSchemaAllTypes(testUserVals),
			},
			{
				ResourceName:            "googleworkspace_user.my-new-user",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccResourceUser_customSchemasMultiple(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	testUserVals := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceUser_customSchemaMultiple(testUserVals),
			},
			{
				ResourceName:            "googleworkspace_user.my-new-user",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func testAccResourceUser_basic(testUserVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_user" "my-new-user" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}
`, testUserVals)
}

func testAccResourceUser_noPassword(testUserVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_user" "my-new-user" {
  primary_email = "%{userEmail}@%{domainName}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}
`, testUserVals)
}

func testAccResourceUser_full(testUserVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_schema" "my-schema" {
  schema_name = "%{userEmail}-schema"

  fields {
    field_name = "birthday"
    field_type = "DATE"
  }

  fields {
    field_name = "favorite-numbers"
    field_type = "INT64"
    multi_valued = true

    numeric_indexing_spec {
      min_value = 1
      max_value = 100
    }
  }
}

resource "googleworkspace_user" "my-new-user" {
  primary_email = "%{userEmail}@%{domainName}"
  password      = "34819d7beeabb9260a5c854bc85b3e44"
  hash_function = "MD5"

  name {
    family_name = "Schrute"
    given_name = "Dwight"
  }

  timeouts {
    create = "15m"
    update = "15m"
  }

  aliases = ["%{userEmail}-assistant_to_regional_manager@%{domainName}", "%{userEmail}-regional_manager@%{domainName}"]

  emails {
    address = "dwight.schrute@example.com"
    type = "home"
  }

  emails {
    address = "dwight.schrute.dunder.mifflin@example.com"
    type = "work"
  }

  external_ids {
    custom_type = "employee_number"
    type = "custom"
    value = "2"
  }

  relations {
    type = "assistant"
    value = "Michael Scott"
  }

  addresses {
    country = "USA"
    country_code = "US"
    locality = "Scranton"
    po_box = "123"
    postal_code = "18508"
    region = "PA"
    street_address = "123 Dunder Mifflin Pkwy"
    type = "work"
  }

  addresses {
    country = "USA"
    country_code = "US"
    locality = "Scranton"
    postal_code = "18508"
    primary = true
    region = "PA"
    street_address = "123 Schrute Farms Rd"
    type = "home"
  }

  organizations {
    department = "sales"
    location = "Scranton"
    name = "Dunder Mifflin"
    primary = true
    symbol = "DUMI"
    title = "member"
    type = "work"
  }

  phones {
    type = "home"
    value = "555-123-7890"
  }

  phones {
    type = "work"
    primary = true
    value = "555-123-0987"
  }

  languages {
    language_code = "en"
  }

  websites {
    primary = true
    type = "blog"
    value = "dundermifflinschrutebeetfarms.blogspot.com"
  }

  locations {
    area = "Scranton"
    building_id = "123"
    desk_code = "1"
    floor_name = "2"
    floor_section = "A"
    type ="desk"
  }

  keywords {
    type = "occupation"
    value = "salesperson"
  }

  ims {
    im = "dwightkschrute"
    primary = true
    protocol = "aim"
    type = "home"
  }

  custom_schemas {
    schema_name = googleworkspace_schema.my-schema.schema_name

    schema_values = {
      "birthday" = jsonencode("1970-01-20")
      "favorite-numbers" = jsonencode([1, 2, 3])
    }
  }

  include_in_global_address_list = false
  recovery_email = "dwightkschrute@example.com"
  recovery_phone = "+16506661212"
  change_password_at_next_login = true
  ip_allowlist = true
}
`, testUserVals)
}

func testAccResourceUser_fullUpdate(testUserVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_schema" "my-schema" {
  schema_name = "%{userEmail}-schema"

  fields {
    field_name = "birthday"
    field_type = "DATE"
  }

  fields {
    field_name = "favorite-numbers"
    field_type = "INT64"
    multi_valued = true

    numeric_indexing_spec {
      min_value = 1
      max_value = 100
    }
  }

  fields {
    field_name = "num-cats"
    field_type = "INT64"
  }
}

resource "googleworkspace_user" "my-new-user" {
  primary_email = "%{userEmail}@%{domainName}"
  password      = "34819d7beeabb9260a5c854bc85b3e44"
  hash_function = "MD5"

  name {
    family_name = "Schrute"
    given_name = "Dwight K"
  }

  timeouts {
    create = "15m"
    update = "15m"
  }

  aliases = ["%{userEmail}-assist_to_regional_manager@%{domainName}"]

  emails {
    address = "dwight.schrute@example.com"
    type = "home"
  }

  external_ids {
    custom_type = "employee_no"
    type = "custom"
    value = "2"
  }

  relations {
    type = "assistant"
    value = "Michael Scott"
  }

  addresses {
    country = "USA"
    country_code = "US"
    locality = "Scranton"
    po_box = "123"
    postal_code = "18508"
    region = "PA"
    street_address = "123 Dunder Mifflin Pkwy"
    type = "work"
  }

  organizations {
    department = "sales"
    location = "Scranton"
    name = "Dunder Mifflin"
    symbol = "DUMI"
    title = "member"
    type = "work"
  }

  phones {
    type = "home"
    value = "555-123-7890"
    primary = true
  }

  languages {
    language_code = "en"
  }

  ssh_public_keys {
    key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPM4pxpbPpjuBocS6qlW0BHRYgH5xmv/yVrANZR9lc1N"
  }

  websites {
    primary = false
    type = "blog"
    value = "dundermifflinschrutebeetfarms.blogspot.com"
  }

  locations {
    area = "Scranton"
    building_id = "123"
    desk_code = "1"
    floor_name = "2"
    floor_section = "B"
    type ="desk"
  }

  keywords {
    type = "occupation"
    value = "salesperson"
  }

  ims {
    im = "dwightkschrute"
    primary = false
    protocol = "aim"
    type = "home"
  }

  custom_schemas {
    schema_name = googleworkspace_schema.my-schema.schema_name

    schema_values = {
      "birthday" = jsonencode("1970-01-20")
      "favorite-numbers" = jsonencode([1, 2])
      "num-cats" = jsonencode(3)
    }
  }

  include_in_global_address_list = true
  recovery_email = "dwight.schrute@example.com"
  recovery_phone = "+16506661212"
  change_password_at_next_login = false
  ip_allowlist = false
}
`, testUserVals)
}

func testAccResourceUser_isAdmin(testUserVals map[string]interface{}, isAdmin string) string {
	testUserVals["isAdmin"] = isAdmin

	return Nprintf(`
resource "googleworkspace_user" "my-new-user" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }

  is_admin = %{isAdmin}
}
`, testUserVals)
}

func testAccResourceUser_suspended(testUserVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_user" "my-new-user" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }

  timeouts {
    update = "15m"
  }

  suspended = true
}
`, testUserVals)
}

// We need to buy Archived User Licenses in order to test this
//func testAccResourceUser_archived(testUserVals map[string]interface{}) string {
//	return Nprintf(`
//resource "googleworkspace_user" "my-new-user" {
//  primary_email = "%{userEmail}@%{domainName}"
//  password = "%{password}"
//
//  name {
//    family_name = "Scott"
//    given_name = "Michael"
//  }
//
//  archived = true
//}
//`, testUserVals)
//}

func testAccResourceUser_customSchemaAllTypes(testUserVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_schema" "my-schema" {
  schema_name = "%{userEmail}-schema"

  fields {
    field_name = "birthday"
    field_type = "DATE"
  }

  fields {
    field_name = "favorite-numbers"
    field_type = "INT64"
    multi_valued = true

    numeric_indexing_spec {
      min_value = 1
      max_value = 100
    }
  }

  fields {
    field_name = "my-custom-phones"
    field_type = "PHONE"
    multi_valued = true
  }

  fields {
    field_name = "my-custom-email"
    field_type = "EMAIL"
  }

  fields {
    field_name = "lbs-of-beets"
    field_type = "DOUBLE"
    multi_valued = true
  }

  fields {
    field_name = "favorite-animal"
    field_type = "STRING"
  }

  fields {
    field_name = "fire-certified"
    field_type = "BOOL"
  }
}

resource "googleworkspace_user" "my-new-user" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }

  custom_schemas {
    schema_name = googleworkspace_schema.my-schema.schema_name

    schema_values = {
      "birthday" = jsonencode("1970-01-20")
      "favorite-numbers" = jsonencode([1, 2, 3])
      "my-custom-phones" = jsonencode(["555-555-5555", "123-456-7890"])
      "my-custom-email" = jsonencode("beets4life@example.com")
      "lbs-of-beets" = jsonencode([1004.35, 100.0, 658])
      "favorite-animal" = jsonencode("bears")
      "fire-certified" =  jsonencode(true)
    }
  }
}
`, testUserVals)
}

func testAccResourceUser_customSchemaMultiple(testUserVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_schema" "bar-schema" {
  schema_name = "%{userEmail}-bar-schema"

  fields {
    field_name = "bar"
    field_type = "STRING"
  }
}

resource "googleworkspace_schema" "baz-schema" {
  schema_name = "%{userEmail}-baz-schema"

  fields {
    field_name = "baz"
    field_type = "STRING"
  }
}

resource "googleworkspace_user" "my-new-user" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }

  custom_schemas {
    schema_name = googleworkspace_schema.bar-schema.schema_name

    schema_values = {
      "bar" = jsonencode("Bar")
    }
  }

  custom_schemas {
    schema_name = googleworkspace_schema.baz-schema.schema_name

    schema_values = {
      "baz" = jsonencode("Baz")
    }
  }
}
`, testUserVals)
}
