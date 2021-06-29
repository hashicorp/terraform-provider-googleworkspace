package googleworkspace

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceGroupSettings_basic(t *testing.T) {
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
				Config: testAccResourceGroupSettings_basic(testGroupVals),
			},
			{
				ResourceName:      "googleworkspace_group_settings.my-group-settings",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceGroupSettings_full(t *testing.T) {
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
				Config: testAccResourceGroupSettings_full(testGroupVals),
			},
			{
				ResourceName:      "googleworkspace_group_settings.my-group-settings",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceGroupSettings_fullUpdate(testGroupVals),
			},
			{
				ResourceName:      "googleworkspace_group.my-group",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceGroupSettings_archive(t *testing.T) {
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
				Config: testAccResourceGroupSettings_archived(testGroupVals),
			},
			{
				ResourceName:      "googleworkspace_group_settings.my-group-settings",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceGroupSettings_unarchived(testGroupVals),
			},
			{
				ResourceName:      "googleworkspace_group.my-group",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceGroupSettings_archiveOnly(testGroupVals),
			},
			{
				ResourceName:      "googleworkspace_group.my-group",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceGroupSettings_basic(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{email}@%{domainName}"
}

resource "googleworkspace_group_settings" "my-group-settings" {
  email = googleworkspace_group.my-group.email
}
`, testGroupVals)
}

func testAccResourceGroupSettings_full(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{email}@%{domainName}"
}

resource "googleworkspace_group_settings" "my-group-settings" {
  email = googleworkspace_group.my-group.email

  allow_external_members = true
  allow_web_posting = false
  is_archived = false
  archive_only = false
  include_custom_footer = true
  send_message_deny_notification = true
  members_can_post_as_the_group = true
  include_in_global_address_list = false
  enable_collaborative_inbox = true

  primary_language = "en"
  custom_reply_to = "my-custom@example.com"
  custom_footer_text = "my-custom-footer"
  default_message_deny_notification_text = "message denied"

  who_can_join = "INVITED_CAN_JOIN"
  who_can_view_membership = "ALL_MANAGERS_CAN_VIEW"
  who_can_view_group = "ALL_MANAGERS_CAN_VIEW"
  who_can_post_message = "ALL_MEMBERS_CAN_POST"
  message_moderation_level = "MODERATE_NEW_MEMBERS"
  spam_moderation_level = "SILENTLY_MODERATE"
  reply_to = "REPLY_TO_CUSTOM"
  who_can_leave_group = "ALL_MANAGERS_CAN_LEAVE"
  who_can_contact_owner = "ALL_MANAGERS_CAN_CONTACT"
  who_can_moderate_members = "NONE"
  who_can_moderate_content = "NONE"
  who_can_assist_content = "OWNERS_ONLY"
  who_can_discover_group = "ALL_MEMBERS_CAN_DISCOVER"
}
`, testGroupVals)
}

func testAccResourceGroupSettings_fullUpdate(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{email}@%{domainName}"
}

resource "googleworkspace_group_settings" "my-group-settings" {
  email = googleworkspace_group.my-group.email

  allow_external_members = false
  allow_web_posting = true
  include_custom_footer = true
  send_message_deny_notification = true
  members_can_post_as_the_group = false

  primary_language = "de"
  custom_reply_to = "my-custom-email@example.com"
  custom_footer_text = "my-custom-footer-update"
  default_message_deny_notification_text = "message denied - updated"

  who_can_join = "ALL_IN_DOMAIN_CAN_JOIN"
  who_can_view_membership = "ALL_IN_DOMAIN_CAN_VIEW"
  who_can_post_message = "ANYONE_CAN_POST"
  message_moderation_level = "MODERATE_ALL_MESSAGES"
  spam_moderation_level = "REJECT"
  reply_to = "REPLY_TO_CUSTOM"
  who_can_leave_group = "NONE_CAN_LEAVE"
  who_can_contact_owner = "ANYONE_CAN_CONTACT"
  who_can_moderate_members = "OWNERS_AND_MANAGERS"
  who_can_moderate_content = "ALL_MEMBERS"
  who_can_assist_content = "OWNERS_AND_MANAGERS"
  who_can_discover_group = "ANYONE_CAN_DISCOVER"
}
`, testGroupVals)
}

func testAccResourceGroupSettings_archived(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{email}@%{domainName}"
}

resource "googleworkspace_group_settings" "my-group-settings" {
  email = googleworkspace_group.my-group.email

  timeouts {
    create = "15m"
    update = "15m"
  }

  is_archived = true
}
`, testGroupVals)
}

func testAccResourceGroupSettings_unarchived(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{email}@%{domainName}"
}

resource "googleworkspace_group_settings" "my-group-settings" {
  email = googleworkspace_group.my-group.email

  is_archived = false
}
`, testGroupVals)
}

func testAccResourceGroupSettings_archiveOnly(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{email}@%{domainName}"
}

resource "googleworkspace_group_settings" "my-group-settings" {
  email = googleworkspace_group.my-group.email

  timeouts {
    create = "10m"
    update = "10m"
  }

  is_archived = true
  archive_only = true
}
`, testGroupVals)
}
