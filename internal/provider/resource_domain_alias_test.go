package googleworkspace

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceDomainAlias(t *testing.T) {

	domainName := os.Getenv("GOOGLEWORKSPACE_DOMAIN")

	if domainName == "" {
		t.Skip("GOOGLEWORKSPACE_DOMAIN needs to be set to run this test")
	}

	domainAlias := fmt.Sprintf("tf-test-%s.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDomainAlias(domainName, domainAlias),
			},
			{
				ResourceName:      "googleworkspace_domain_alias.my-domain-alias",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceDomainAlias(domainName, domainAlias string) string {
	return fmt.Sprintf(`
resource "googleworkspace_domain_alias" "my-domain-alias" {
  parent_domain_name = "%s"
  domain_alias_name = "%s"
}
`, domainName, domainAlias)
}
