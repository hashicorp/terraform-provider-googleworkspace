package googleworkspace

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceUsers(t *testing.T) {
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
				Config: testAccDataSourceUsers(testUserVals),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.googleworkspace_users.users",
						"users.#"),
					resource.TestMatchResourceAttr("data.googleworkspace_users.users",
						"users.0.primary_email", regexp.MustCompile(fmt.Sprintf("^.*@%s$", domainName))),
				),
			},
		},
	})
}

func testAccDataSourceUsers(testUserVals map[string]interface{}) string {
	return testAccResourceUser_basic(testUserVals) + `

data "googleworkspace_users" "users" {
  depends_on = [googleworkspace_user.my-new-user]
}
`
}
