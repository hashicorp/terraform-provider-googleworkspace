package googleworkspace

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceGroupMembers_basic(t *testing.T) {
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
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccResourceGroupMembersExists("googleworkspace_group_members.my-group-members"),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGroupMembers_basic(testGroupVals),
			},
			{
				ResourceName:      "googleworkspace_group_members.my-group-members",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceGroupMembers_full(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	testGroupVals := map[string]interface{}{
		"domainName": domainName,
		"userEmail1": fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"userEmail2": fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"groupEmail": fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccResourceGroupMembersExists("googleworkspace_group_members.my-group-members"),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGroupMembers_full(testGroupVals),
			},
			{
				ResourceName:      "googleworkspace_group_members.my-group-members",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceGroupMembers_fullUpdate(testGroupVals),
			},
			{
				ResourceName:      "googleworkspace_group_members.my-group-members",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceGroupMembersExists(resource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("%s key not found in state", resource)
		}

		client, err := googleworkspaceTestClient()
		if err != nil {
			return err
		}

		directoryService, diags := client.NewDirectoryService()
		if diags.HasError() {
			return fmt.Errorf("Error creating directory service %+v", diags)
		}

		membersService, diags := GetMembersService(directoryService)
		if diags.HasError() {
			return fmt.Errorf("Error getting group members service %+v", diags)
		}

		parts := strings.Split(rs.Primary.ID, "/")

		// id is of format "groups/<group_id>"
		if len(parts) != 2 {
			return fmt.Errorf("Group Members Id (%s) is not of the correct format (groups/<group_id>)", rs.Primary.ID)
		}

		members, err := membersService.List(parts[1]).Do()
		if err == nil && len(members.Members) > 0 {
			return fmt.Errorf("Group Members still exists (%s)", rs.Primary.ID)
		}

		return nil
	}
}

func testAccResourceGroupMembers_basic(testGroupVals map[string]interface{}) string {
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

resource "googleworkspace_group_members" "my-group-members" {
  group_id = googleworkspace_group.my-group.id

	members {
		email = googleworkspace_user.my-new-user.primary_email
	}
}
`, testGroupVals)
}

func testAccResourceGroupMembers_full(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{groupEmail}@%{domainName}"
}

resource "googleworkspace_user" "my-new-user1" {
  primary_email = "%{userEmail1}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_user" "my-new-user2" {
  primary_email = "%{userEmail2}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_group_members" "my-group-members" {
  group_id = googleworkspace_group.my-group.id

	members {
		email = googleworkspace_user.my-new-user1.primary_email
		role = "MANAGER"
		type = "USER"
		delivery_settings = "ALL_MAIL"
	}

	members {
		email = googleworkspace_user.my-new-user2.primary_email
		role = "MANAGER"
		type = "USER"
		delivery_settings = "ALL_MAIL"
	}
}
`, testGroupVals)
}

func testAccResourceGroupMembers_fullUpdate(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{groupEmail}@%{domainName}"
}

resource "googleworkspace_user" "my-new-user1" {
  primary_email = "%{userEmail1}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_user" "my-new-user2" {
  primary_email = "%{userEmail2}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_group_members" "my-group-members" {
  group_id = googleworkspace_group.my-group.id

	members {
		email = googleworkspace_user.my-new-user1.primary_email
		role = "OWNER"
		type = "USER"
		delivery_settings = "DAILY"
	}

	members {
		email = googleworkspace_user.my-new-user2.primary_email
		role = "MANAGER"
		type = "USER"
		delivery_settings = "ALL_MAIL"
	}
}
`, testGroupVals)
}
