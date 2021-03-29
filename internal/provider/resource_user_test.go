package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceUser_basic(t *testing.T) {
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

func TestAccResourceUser_full(t *testing.T) {
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
				Config: testAccResourceUser_full(testUserVals),
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

func testAccResourceUser_full(testUserVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_user" "my-new-user" {
  primary_email = "%{userEmail}@%{domainName}"
  password      = "%{password}"

  name {
    family_name = "Schrute"
    given_name = "Dwight"
  }

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

  recovery_email = "dwightkschrute@example.com"
  recovery_phone = "+16506661212"
  change_password_at_next_login = true
  ip_whitelisted = true
}
`, testUserVals)
}
