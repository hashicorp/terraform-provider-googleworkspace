// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"google.golang.org/api/chromepolicy/v1"
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

func TestAccResourceChromePolicy_typeMessage(t *testing.T) {
	t.Parallel()

	ouName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceChromePolicy_typeMessage(ouName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.#", "1"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.0.schema_name", "chrome.users.ManagedBookmarksSetting"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.0.schema_values.managedBookmarks", "{\"toplevelName\":\"Stuff\"}"),
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

	// ensures previously set field was reset/removed
	// this passing also implies Delete works correctly
	// based on the implementation
	testCheck := func(s *terraform.State) error {
		client, err := googleworkspaceTestClient()
		if err != nil {
			return err
		}

		rs, ok := s.RootModule().Resources["googleworkspace_org_unit.test"]
		if !ok {
			return fmt.Errorf("Can't find org unit resource: googleworkspace_org_unit.test")
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("org unit ID not set")
		}

		chromePolicyService, diags := client.NewChromePolicyService()
		if diags.HasError() {
			return errors.New(diags[0].Summary)
		}

		chromePoliciesService, diags := GetChromePoliciesService(chromePolicyService)
		if diags.HasError() {
			return errors.New(diags[0].Summary)
		}

		policyTargetKey := &chromepolicy.GoogleChromePolicyV1PolicyTargetKey{
			TargetResource: "orgunits/" + strings.TrimPrefix(rs.Primary.ID, "id:"),
		}

		resp, err := chromePoliciesService.Resolve(fmt.Sprintf("customers/%s", client.Customer), &chromepolicy.GoogleChromePolicyV1ResolveRequest{
			PolicySchemaFilter: "chrome.users.MaxConnectionsPerProxy",
			PolicyTargetKey:    policyTargetKey,
		}).Do()
		if err != nil {
			return err
		}
		if len(resp.ResolvedPolicies) > 0 {
			return fmt.Errorf("Expected policy to be reset")
		}
		return nil
	}

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
				Config: testAccResourceChromePolicy_multipleRearranged(ouName, 34, ".*@example.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.#", "2"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.0.schema_name", "chrome.users.MaxConnectionsPerProxy"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.0.schema_values.maxConnectionsPerProxy", "34"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.1.schema_name", "chrome.users.RestrictSigninToPattern"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.1.schema_values.restrictSigninToPattern", encode(".*@example.com")),
				),
			},
			{
				Config: testAccResourceChromePolicy_multipleDifferent(ouName, true, ".*@example.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.#", "2"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.0.schema_name", "chrome.users.OnlineRevocationChecks"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.0.schema_values.enableOnlineRevocationChecks", "true"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.1.schema_name", "chrome.users.RestrictSigninToPattern"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.1.schema_values.restrictSigninToPattern", encode(".*@example.com")),
					testCheck,
				),
			},
			{
				Config: testAccResourceChromePolicy_multipleValueTypes(ouName, true, "POLICY_MODE_RECOMMENDED"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.#", "1"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.0.schema_name", "chrome.users.DomainReliabilityAllowed"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.0.schema_values.domainReliabilityAllowed", "true"),
					resource.TestCheckResourceAttr("googleworkspace_chrome_policy.test", "policies.0.schema_values.domainReliabilityAllowedSettingGroupPolicyMode", encode("POLICY_MODE_RECOMMENDED")),
					testCheck,
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

func testAccResourceChromePolicy_multipleRearranged(ouName string, conns int, pattern string) string {
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
  policies {
    schema_name = "chrome.users.RestrictSigninToPattern"
    schema_values = {
      restrictSigninToPattern = jsonencode("%s")
    }
  }
}
`, ouName, conns, pattern)
}

func testAccResourceChromePolicy_multipleDifferent(ouName string, enabled bool, pattern string) string {
	return fmt.Sprintf(`
resource "googleworkspace_org_unit" "test" {
  name = "%s"
  parent_org_unit_path = "/"
}

resource "googleworkspace_chrome_policy" "test" {
  org_unit_id = googleworkspace_org_unit.test.id
  policies {
    schema_name = "chrome.users.OnlineRevocationChecks"
    schema_values = {
      enableOnlineRevocationChecks = jsonencode(%t)
    }
  }
  policies {
    schema_name = "chrome.users.RestrictSigninToPattern"
    schema_values = {
      restrictSigninToPattern = jsonencode("%s")
    }
  }
}
`, ouName, enabled, pattern)
}

func testAccResourceChromePolicy_multipleValueTypes(ouName string, enabled bool, policyMode string) string {
	return fmt.Sprintf(`
resource "googleworkspace_org_unit" "test" {
  name = "%s"
  parent_org_unit_path = "/"
}

resource "googleworkspace_chrome_policy" "test" {
  org_unit_id = googleworkspace_org_unit.test.id
  policies {
    schema_name = "chrome.users.DomainReliabilityAllowed"
    schema_values = {
	  domainReliabilityAllowed                       = jsonencode(%t)
      domainReliabilityAllowedSettingGroupPolicyMode = jsonencode("%s")
    }
  }
}
`, ouName, enabled, policyMode)
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

func testAccResourceChromePolicy_typeMessage(ouName string) string {
	return fmt.Sprintf(`
resource "googleworkspace_org_unit" "test" {
  name = "%s"
  parent_org_unit_path = "/"
}

resource "googleworkspace_chrome_policy" "test" {
  org_unit_id = googleworkspace_org_unit.test.id
  policies {
    schema_name = "chrome.users.ManagedBookmarksSetting"
    schema_values = {
		managedBookmarks = "{\"toplevelName\":\"Stuff\"}"
    }
  }
}
`, ouName)
}
