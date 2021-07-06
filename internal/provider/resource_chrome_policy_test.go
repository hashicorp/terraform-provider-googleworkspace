package googleworkspace

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceChromePolicy_basic(t *testing.T) {
	t.Parallel()

	ouName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceChromePolicy_basic(ouName, 33),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.#", "1"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.0.schema_name", "chrome.users.MaxConnectionsPerProxy"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.0.schema_values.maxConnectionsPerProxy", "33"),
				),
			},
		},
	})
}

func TestAccResourceChromePolicy_update(t *testing.T) {
	t.Parallel()

	ouName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceChromePolicy_basic(ouName, 33),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.#", "1"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.0.schema_name", "chrome.users.MaxConnectionsPerProxy"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.0.schema_values.maxConnectionsPerProxy", "33"),
				),
			},
			{
				Config: testAccResourceChromePolicy_basic(ouName, 34),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.#", "1"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.0.schema_name", "chrome.users.MaxConnectionsPerProxy"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.0.schema_values.maxConnectionsPerProxy", "34"),
				),
			},
		},
	})
}

func TestAccResourceChromePolicy_multiple(t *testing.T) {
	t.Parallel()

	ouName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceChromePolicy_multiple(ouName, 33, ".*@example"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.#", "2"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.0.schema_name", "chrome.users.RestrictSigninToPattern"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.0.schema_values.restrictSigninToPattern", encode(".*@example")),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.1.schema_name", "chrome.users.MaxConnectionsPerProxy"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.1.schema_values.maxConnectionsPerProxy", "33"),
				),
			},
			{
				Config: testAccResourceChromePolicy_multiple(ouName, 34, ".*@example.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.#", "2"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.0.schema_name", "chrome.users.RestrictSigninToPattern"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.0.schema_values.restrictSigninToPattern", encode(".*@example.com")),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.1.schema_name", "chrome.users.MaxConnectionsPerProxy"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.1.schema_values.maxConnectionsPerProxy", "34"),
				),
			},
		},
	})
}

func encode(content string) string {
	res, _ := json.Marshal(content)
	return string(res)
}

func testAccResourceChromePolicy_multiple(ouName string, conns int, pattern string) string {
	return fmt.Sprintf(`
resource "googleworkspace_org_unit" "test" {
  name = "%s"
  parent_org_unit_path = "/"
}

resource "googleworkspace_chrome_policy" "test" {
  org_unit_id = googleworkspace_org_unit.test.id
  policies {
    schema_name = "chrome.users.RestrictSigninToPattern"
    schema_values = {
      restrictSigninToPattern = jsonencode("%s")
    }
  }
  policies {
    schema_name = "chrome.users.MaxConnectionsPerProxy"
    schema_values = {
      maxConnectionsPerProxy = jsonencode(%d)
    }
  }
}
`, ouName, pattern, conns)
}

func testAccResourceChromePolicy_basic(ouName string, conns int) string {
	return fmt.Sprintf(`
resource "googleworkspace_org_unit" "test" {
  name = "%s"
  parent_org_unit_path = "/"
}

resource "googleworkspace_chrome_policy" "test" {
  org_unit_id = googleworkspace_org_unit.test.id
  policies {
    schema_name = "chrome.users.MaxConnectionsPerProxy"
    schema_values = {
      maxConnectionsPerProxy = jsonencode(%d)
    }
  }
}
`, ouName, conns)
}
