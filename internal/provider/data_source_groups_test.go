package googleworkspace

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceGroups(t *testing.T) {
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceGroups(testGroupVals),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.googleworkspace_groups.groups",
						"groups.#"),
					resource.TestMatchResourceAttr("data.googleworkspace_groups.groups",
						"groups.0.email", regexp.MustCompile(fmt.Sprintf("^.*@%s$", domainName))),
				),
			},
		},
	})
}

func testAccDataSourceGroups(testUserVals map[string]interface{}) string {
	return testAccResourceGroup_full(testUserVals) + `

data "googleworkspace_groups" "groups" {
  depends_on = [googleworkspace_group.my-group]
}
`
}
