// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDomain() *schema.Resource {
	dsSchema := datasourceSchemaFromResourceSchema(resourceDomain().Schema)
	addRequiredFieldsToSchema(dsSchema, "domain_name")

	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Domain data source in the Terraform Googleworkspace provider. Domain resides under the " +
			"`https://www.googleapis.com/auth/admin.directory.domain` client scope.",

		ReadContext: dataSourceDomainRead,

		Schema: dsSchema,
	}
}

func dataSourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId(d.Get("domain_name").(string))
	return resourceDomainRead(ctx, d, meta)
}
