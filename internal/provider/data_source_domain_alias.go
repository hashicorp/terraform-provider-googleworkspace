// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDomainAlias() *schema.Resource {
	dsSchema := datasourceSchemaFromResourceSchema(resourceDomainAlias().Schema)
	addRequiredFieldsToSchema(dsSchema, "domain_alias_name")

	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Domain Alias data source in the Terraform Googleworkspace provider. Domain Alias resides under the " +
			"`https://www.googleapis.com/auth/admin.directory.domain` client scope.",

		ReadContext: dataSourceDomainAliasRead,

		Schema: dsSchema,
	}
}

func dataSourceDomainAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId(d.Get("domain_alias_name").(string))
	return resourceDomainAliasRead(ctx, d, meta)
}
