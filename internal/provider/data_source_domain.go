package googleworkspace

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDomain() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Domain data source in the Terraform Googleworkspace provider.",

		ReadContext: dataSourceDomainRead,

		Schema: resourceDomain().Schema,
	}
}

func dataSourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId(d.Get("domain_name").(string))
	return resourceDomainRead(ctx, d, meta)
}
