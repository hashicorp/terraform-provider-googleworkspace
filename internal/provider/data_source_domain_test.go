// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceDomain(t *testing.T) {
	domainName := fmt.Sprintf("tf-test-%s.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDomain(domainName),
				Check:  resource.TestCheckResourceAttr("data.googleworkspace_domain.my-domain", "domain_name", domainName),
			},
		},
	})
}

func testAccDataSourceDomain(domainName string) string {
	return fmt.Sprintf(`
resource "googleworkspace_domain" "my-domain" {
  domain_name = "%s"
}

data "googleworkspace_domain" "my-domain" {
  domain_name = googleworkspace_domain.my-domain.domain_name
}
`, domainName)
}
