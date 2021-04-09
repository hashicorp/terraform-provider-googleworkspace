package googleworkspace

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceUser_withId(t *testing.T) {
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
				Config: testAccDataSourceUser_withId(testUserVals),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.googleworkspace_user.my-new-user", "primary_email", Nprintf("%{userEmail}@%{domainName}", testUserVals)),
					resource.TestCheckResourceAttr(
						"data.googleworkspace_user.my-new-user", "name.0.family_name", "Scott"),
					resource.TestCheckResourceAttr(
						"data.googleworkspace_user.my-new-user", "name.0.given_name", "Michael"),
					resource.TestCheckResourceAttr(
						"data.googleworkspace_user.my-new-user", "emails.#", "2"),
				),
			},
		},
	})
}

func TestAccDataSourceUser_withEmail(t *testing.T) {
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
				Config: testAccDataSourceUser_withEmail(testUserVals),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.googleworkspace_user.my-new-user", "primary_email", Nprintf("%{userEmail}@%{domainName}", testUserVals)),
					resource.TestCheckResourceAttr(
						"data.googleworkspace_user.my-new-user", "name.0.family_name", "Scott"),
					resource.TestCheckResourceAttr(
						"data.googleworkspace_user.my-new-user", "name.0.given_name", "Michael"),
					resource.TestCheckResourceAttr(
						"data.googleworkspace_user.my-new-user", "emails.#", "2"),
				),
			},
		},
	})
}

func testAccDataSourceUser_withId(testUserVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_user" "my-new-user" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

data "googleworkspace_user" "my-new-user" {
  id = googleworkspace_user.my-new-user.id
}
`, testUserVals)
}

func testAccDataSourceUser_withEmail(testUserVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_user" "my-new-user" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

data "googleworkspace_user" "my-new-user" {
  primary_email = googleworkspace_user.my-new-user.primary_email
}
`, testUserVals)
}
