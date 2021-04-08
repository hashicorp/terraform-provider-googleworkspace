package googleworkspace

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceGroup_basic(t *testing.T) {
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
				Config: testAccResourceGroup_basic(testGroupVals),
			},
			{
				ResourceName:      "googleworkspace_group.my-group",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceGroup_full(t *testing.T) {
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
				Config: testAccResourceGroup_full(testGroupVals),
			},
			{
				ResourceName:      "googleworkspace_group.my-group",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceGroup_fullUpdate(testGroupVals),
			},
			{
				ResourceName:      "googleworkspace_group.my-group",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceGroup_basic(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{email}@%{domainName}"
}
`, testGroupVals)
}

func testAccResourceGroup_full(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{email}@%{domainName}"
  name  = "tf-test-name"
  description = "my test description"

  aliases = ["%{email}-alias-1@%{domainName}", "%{email}-alias-2@%{domainName}"]
}
`, testGroupVals)
}

func testAccResourceGroup_fullUpdate(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{email}@%{domainName}"
  name  = "tf-new-name"
  description = "my new description"

  aliases = ["%{email}-alias-2@%{domainName}", "%{email}-new-alias@%{domainName}"]
}
`, testGroupVals)
}
