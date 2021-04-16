package googleworkspace

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceGroupMember_basic(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	testGroupVals := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"groupEmail": fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGroupMember_basic(testGroupVals),
			},
			{
				ResourceName:      "googleworkspace_group_member.my-group-member",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceGroupMember_full(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	testGroupVals := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"groupEmail": fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGroupMember_full(testGroupVals),
			},
			{
				ResourceName:      "googleworkspace_group_member.my-group-member",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceGroupMember_fullUpdate(testGroupVals),
			},
			{
				ResourceName:      "googleworkspace_group_member.my-group-member",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceGroupMember_basic(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{groupEmail}@%{domainName}"
}

resource "googleworkspace_user" "my-new-user" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_group_member" "my-group-member" {
  group_id = googleworkspace_group.my-group.id
  email = googleworkspace_user.my-new-user.primary_email
}
`, testGroupVals)
}

func testAccResourceGroupMember_full(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{groupEmail}@%{domainName}"
}

resource "googleworkspace_user" "my-new-user" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_group_member" "my-group-member" {
  group_id = googleworkspace_group.my-group.id
  email = googleworkspace_user.my-new-user.primary_email

  role = "MANAGER"
  type = "USER"
  delivery_settings = "ALL_MAIL"
}
`, testGroupVals)
}

func testAccResourceGroupMember_fullUpdate(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{groupEmail}@%{domainName}"
}

resource "googleworkspace_user" "my-new-user" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_group_member" "my-group-member" {
  group_id = googleworkspace_group.my-group.id
  email = googleworkspace_user.my-new-user.primary_email

  role = "OWNER"
  type = "USER"
  delivery_settings = "DAILY"
}
`, testGroupVals)
}
