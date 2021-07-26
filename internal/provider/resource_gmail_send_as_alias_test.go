package googleworkspace

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceGmailSendAsAlias_basic(t *testing.T) {
	t.Parallel()

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}
	data := map[string]interface{}{
		"domainName": domainName,
		"userEmail":  fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		"password":   acctest.RandString(10),
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

func testAccGmailSendAsAlias_basic(data map[string]interface{}) string {
	return Nprintf(`
resource "googleworkspace_user" "test" {
  primary_email = "%{userEmail}@%{domainName}"
  password = "%{password}"

  name {
    family_name = "Scott"
    given_name = "Michael"
  }
}

resource "googleworkspace_gmail_send_as_alias" "test" {
  primary_email = googleworkspace_user.test.primary_email
  send_as_email = "test@test.com"
}
`, data)
}
