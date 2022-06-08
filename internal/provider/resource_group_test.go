package googleworkspace

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

	expectedEmail := fmt.Sprintf("%s@%s", testGroupVals["email"].(string), testGroupVals["domainName"].(string))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGroup_basic(testGroupVals),
			},
			{
				// TestStep imports by `id` by default - an alphanumeric string
				// https://developers.google.com/admin-sdk/directory/v1/guides/manage-groups#get_group
				ResourceName:            "googleworkspace_group.my-group",
				ImportState:             true,
				ImportStateCheck:        checkGroupImportState(),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
			{
				// TestStep imports by `email`
				// https://developers.google.com/admin-sdk/directory/v1/guides/manage-groups#get_group
				ResourceName:            "googleworkspace_group.my-group",
				ImportState:             true,
				ImportStateId:           expectedEmail,
				ImportStateCheck:        checkGroupImportState(),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
		},
	})
}

func checkGroupImportState() resource.ImportStateCheckFunc {
	return resource.ImportStateCheckFunc(
		func(state []*terraform.InstanceState) error {
			if len(state) > 1 {
				return fmt.Errorf("state should only contain one group resource, got: %d", len(state))
			}

			id := state[0].ID
			isEmail := isEmail(id)
			if isEmail {
				return fmt.Errorf("id should be random alphanumeric string, got email: %s", id)
			}

			return nil
		},
	)
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
				ResourceName:            "googleworkspace_group.my-group",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
			},
			{
				Config: testAccResourceGroup_fullUpdate(testGroupVals),
			},
			{
				ResourceName:            "googleworkspace_group.my-group",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"etag"},
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
