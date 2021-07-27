package googleworkspace

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceGmailSendAsAlias_basic(t *testing.T) {
	gmailUser := os.Getenv("GOOGLEWORKSPACE_TEST_GMAIL_USER")

	if gmailUser == "" {
		t.Skip("GOOGLEWORKSPACE_TEST_GMAIL_USER needs to be set to run this test")
	}

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	data := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
		"gmailUser":  gmailUser,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGmailSendAsAlias_basic(data),
			},
			{
				ResourceName:      "googleworkspace_gmail_send_as_alias.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceGmailSendAsAlias_full(t *testing.T) {
	gmailUser := os.Getenv("GOOGLEWORKSPACE_TEST_GMAIL_USER")

	if gmailUser == "" {
		t.Skip("GOOGLEWORKSPACE_TEST_GMAIL_USER needs to be set to run this test")
	}

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	data := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
		"gmailUser":  gmailUser,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGmailSendAsAlias_basic(data),
			},
			{
				ResourceName:      "googleworkspace_gmail_send_as_alias.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGmailSendAsAlias_full(data),
			},
			{
				ResourceName:      "googleworkspace_gmail_send_as_alias.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceGmailSendAsAlias_defaultCannotBeSetToFalseDirectly(t *testing.T) {
	gmailUser := os.Getenv("GOOGLEWORKSPACE_TEST_GMAIL_USER")

	if gmailUser == "" {
		t.Skip("GOOGLEWORKSPACE_TEST_GMAIL_USER needs to be set to run this test")
	}

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	data := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
		"gmailUser":  gmailUser,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGmailSendAsAlias_withDefault(data),
			},
			{
				Config:      testAccGmailSendAsAlias_basic(data),
				ExpectError: regexp.MustCompile("isDefault cannot be toggled to false"),
			},
		},
	})
}

func TestAccResourceGmailSendAsAlias_switchingDefaultCausesDiff(t *testing.T) {
	gmailUser := os.Getenv("GOOGLEWORKSPACE_TEST_GMAIL_USER")

	if gmailUser == "" {
		t.Skip("GOOGLEWORKSPACE_TEST_GMAIL_USER needs to be set to run this test")
	}

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	data := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
		"userEmail2": fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"gmailUser":  gmailUser,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGmailSendAsAlias_withDefaultUser1(data),
			},
			{
				Config:             testAccGmailSendAsAlias_withDefaultUser2(data),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGmailSendAsAlias_basic(data map[string]interface{}) string {
	return Nprintf(`
data "googleworkspace_user" "test" {
  primary_email = "%{gmailUser}"
}

resource "googleworkspace_user" "alias" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_gmail_send_as_alias" "test" {
  primary_email = data.googleworkspace_user.test.primary_email
  send_as_email = googleworkspace_user.alias.primary_email
}
`, data)
}

func testAccGmailSendAsAlias_full(data map[string]interface{}) string {
	return Nprintf(`
data "googleworkspace_user" "test" {
  primary_email = "%{gmailUser}"
}

resource "googleworkspace_user" "alias" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_gmail_send_as_alias" "test" {
  primary_email = data.googleworkspace_user.test.primary_email
  send_as_email = googleworkspace_user.alias.primary_email
  display_name = "test"
  reply_to_address = "test@test.com"
  signature = "<b>yours truly</b>"
  treat_as_alias = false
}
`, data)
}

func testAccGmailSendAsAlias_withDefault(data map[string]interface{}) string {
	return Nprintf(`
data "googleworkspace_user" "test" {
  primary_email = "%{gmailUser}"
}

resource "googleworkspace_user" "alias" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_gmail_send_as_alias" "test" {
  primary_email = data.googleworkspace_user.test.primary_email
  send_as_email = googleworkspace_user.alias.primary_email
  is_default = true
}
`, data)
}

func testAccGmailSendAsAlias_withDefaultUser1(data map[string]interface{}) string {
	return Nprintf(`
data "googleworkspace_user" "test" {
  primary_email = "%{gmailUser}"
}

resource "googleworkspace_user" "alias" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_user" "alias2" {
  primary_email = "%{userEmail2}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Schrute"
    given_name = "Dwight"
  }
}

resource "googleworkspace_gmail_send_as_alias" "test" {
  primary_email = data.googleworkspace_user.test.primary_email
  send_as_email = googleworkspace_user.alias.primary_email
  is_default = true
}

resource "googleworkspace_gmail_send_as_alias" "test2" {
  primary_email = data.googleworkspace_user.test.primary_email
  send_as_email = googleworkspace_user.alias2.primary_email
}
`, data)
}

func testAccGmailSendAsAlias_withDefaultUser2(data map[string]interface{}) string {
	return Nprintf(`
data "googleworkspace_user" "test" {
  primary_email = "%{gmailUser}"
}

resource "googleworkspace_user" "alias" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_user" "alias2" {
  primary_email = "%{userEmail2}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Schrute"
    given_name = "Dwight"
  }
}

resource "googleworkspace_gmail_send_as_alias" "test" {
  primary_email = data.googleworkspace_user.test.primary_email
  send_as_email = googleworkspace_user.alias.primary_email
  is_default = true
}

resource "googleworkspace_gmail_send_as_alias" "test2" {
  primary_email = data.googleworkspace_user.test.primary_email
  send_as_email = googleworkspace_user.alias2.primary_email
  is_default = true
}
`, data)
}
