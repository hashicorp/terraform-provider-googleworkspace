package googleworkspace

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceUserAlias_basic(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	testId := acctest.RandString(10)

	testAliasVals := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", testId),
		"alias":      fmt.Sprintf("tf-test-alias-%s", testId),
		"password":   acctest.RandString(16),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserAlias_basic(testAliasVals),
			},
			{
				ResourceName:            "googleworkspace_user_alias.test",
				ImportState:             true,
				ImportStateVerify:       true,
			},
		},
	})
}

func testAccUserAlias_basic(testAliasVars map[string]interface{}) string {
	return Nprintf(`
	resource "googleworkspace_user" "test" {
		primary_email = "%{userEmail}@%{domainName}"
		password = "%{password}"
		name {
			family_name = "User"
			given_name = "Test"
		}

		lifecycle {
			ignore_changes = [aliases]
		  }
	}

	resource "googleworkspace_user_alias" "test" {
		primary_email = googleworkspace_user.test.primary_email
		alias = "%{alias}@%{domainName}"
	}
	`, testAliasVars)
}
