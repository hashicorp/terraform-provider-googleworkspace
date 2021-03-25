package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceDomain(t *testing.T) {

	domainName := fmt.Sprintf("tf-test-%s.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDomain(domainName),
			},
			{
				ResourceName:      "googleworkspace_domain.my-domain",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceDomain(domainName string) string {
	return fmt.Sprintf(`
resource "googleworkspace_domain" "my-domain" {
  domain_name = "%s"
}
`, domainName)
}
