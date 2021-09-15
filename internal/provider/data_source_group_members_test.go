package googleworkspace

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceGroupMembers(t *testing.T) {
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
				Config: testAccDataSourceGroupMembers(testGroupVals),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.googleworkspace_group_members.my-group-members", "members.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.googleworkspace_group_members.my-group-members", "members", map[string]string{
							"email": Nprintf("%{userEmail}@%{domainName}", testGroupVals),
							"role":  "MEMBER",
							"type":  "USER",
						}),
				),
			},
		},
	})
}

func testAccDataSourceGroupMembers(testGroupVals map[string]interface{}) string {
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

data "googleworkspace_group_members" "my-group-members" {
  group_id = googleworkspace_group.my-group.id

  depends_on = [googleworkspace_group_member.my-group-member]
}
`, testGroupVals)
}
