package googleworkspace

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
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
		"userEmail":  fmt.Sprintf("tf-test-%s@%s", acctest.RandString(10), domainName),
		"groupEmail": fmt.Sprintf("tf-test-%s@%s", acctest.RandString(10), domainName),
		"password":   acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccResourceGroupMembersExists("googleworkspace_group_members.my-group-members"),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGroupMembers_basic(testGroupVals),
				Check: testAccCheckGoogleWorkspaceMembers(t, []map[string]interface{}{
					{
						"email": testGroupVals["userEmail"],
						"type":  "USER",
						"role":  "MEMBER",
					},
				}),
			},
			{
				ResourceName: "googleworkspace_group_members.my-group-members",
				ImportState:  true,
			},
		},
	})
}

func TestAccResourceGroupMembers_empty(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	testGroupVals := map[string]interface{}{
		"userEmail":  fmt.Sprintf("tf-test-%s@%s", acctest.RandString(10), domainName),
		"groupEmail": fmt.Sprintf("tf-test-%s@%s", acctest.RandString(10), domainName),
		"password":   acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccResourceGroupMembersExists("googleworkspace_group_members.my-group-members"),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGroupMembers_basic(testGroupVals),
				Check: testAccCheckGoogleWorkspaceMembers(t, []map[string]interface{}{
					{
						"email": testGroupVals["userEmail"],
						"type":  "USER",
						"role":  "MEMBER",
					},
				}),
			},
			{
				ResourceName: "googleworkspace_group_members.my-group-members",
				ImportState:  true,
			},
			{
				Config: testAccResourceGroupMembers_empty(testGroupVals),
				Check:  testAccCheckGoogleWorkspaceMembers(t, []map[string]interface{}{}),
			},
			{
				ResourceName: "googleworkspace_group_members.my-group-members",
				ImportState:  true,
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
		"userEmail1": fmt.Sprintf("tf-test-%s@%s", acctest.RandString(10), domainName),
		"userEmail2": fmt.Sprintf("tf-test-%s@%s", acctest.RandString(10), domainName),
		"groupEmail": fmt.Sprintf("tf-test-%s@%s", acctest.RandString(10), domainName),
		"password":   acctest.RandString(10),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccResourceGroupMembersExists("googleworkspace_group_members.my-group-members"),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGroupMembers_full(testGroupVals),
				Check: testAccCheckGoogleWorkspaceMembers(t, []map[string]interface{}{
					{
						"email": testGroupVals["userEmail1"],
						"type":  "USER",
						"role":  "MANAGER",
					},
					{
						"email": testGroupVals["userEmail2"],
						"type":  "USER",
						"role":  "MANAGER",
					},
				}),
			},
			{
				ResourceName: "googleworkspace_group_members.my-group-members",
				ImportState:  true,
			},
			{
				Config: testAccResourceGroupMembers_fullUpdate(testGroupVals),
				Check: testAccCheckGoogleWorkspaceMembers(t, []map[string]interface{}{
					{
						"email": testGroupVals["userEmail1"],
						"type":  "USER",
						"role":  "OWNER",
					},
					{
						"email": testGroupVals["userEmail2"],
						"type":  "USER",
						"role":  "MANAGER",
					},
				}),
			},
			{
				ResourceName: "googleworkspace_group_members.my-group-members",
				ImportState:  true,
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

		diags := diag.Diagnostics{}
		client := googleworkspaceTestClient(context.Background(), &diags)
		if diags.HasError() {
			return getDiagErrors(diags)
		}

		membersService := GetMembersService(client, &diags)
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

func testAccCheckGoogleWorkspaceMembers(t *testing.T, members []map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["googleworkspace_group_members.my-group-members"]
		if !ok {
			return fmt.Errorf("Resource not found: googleworkspace_group_members.my-group-members")
		}

		for _, mem := range members {
			prefix := ""
			for attrKey, attrVal := range rs.Primary.Attributes {
				if attrVal == mem["email"] {
					prefix = strings.Join(strings.Split(attrKey, ".")[0:2], ".")
					break
				}
			}

			if prefix == "" {
				return fmt.Errorf("No members matching %s", mem["email"])
			}

			if rs.Primary.Attributes[prefix+".role"] != mem["role"] {
				return fmt.Errorf("Member %s does not match role state: %s | expected: %s", mem["email"], rs.Primary.Attributes[prefix+".role"], mem["role"])
			}

			if rs.Primary.Attributes[prefix+".type"] != mem["type"] {
				return fmt.Errorf("Member %s does not match type state: %s | expected: %s", mem["email"], rs.Primary.Attributes[prefix+".type"], mem["type"])
			}
		}

		return nil
	}
}

func testAccResourceGroupMembers_basic(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{groupEmail}"
}

resource "googleworkspace_user" "my-new-user" {
  primary_email = "%{userEmail}"
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

func testAccResourceGroupMembers_empty(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{groupEmail}"
}

resource "googleworkspace_user" "my-new-user" {
  primary_email = "%{userEmail}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_group_members" "my-group-members" {
  group_id = googleworkspace_group.my-group.id
}
`, testGroupVals)
}

func testAccResourceGroupMembers_full(testGroupVals map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_group" "my-group" {
  email = "%{groupEmail}"
}

resource "googleworkspace_user" "my-new-user1" {
  primary_email = "%{userEmail1}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_user" "my-new-user2" {
  primary_email = "%{userEmail2}"
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
  email = "%{groupEmail}"
}

resource "googleworkspace_user" "my-new-user1" {
  primary_email = "%{userEmail1}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_user" "my-new-user2" {
  primary_email = "%{userEmail2}"
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
