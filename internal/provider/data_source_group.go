// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGroup() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceGroup().Schema)
	addExactlyOneOfFieldsToSchema(dsSchema, "id", "email")

	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Group data source in the Terraform Googleworkspace provider. Group resides under the " +
			"`https://www.googleapis.com/auth/admin.directory.group` client scope.",

		ReadContext: dataSourceGroupRead,

		Schema: dsSchema,
	}
}

func dataSourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Get("id") != "" {
		d.SetId(d.Get("id").(string))
	} else {
		var diags diag.Diagnostics

		// use the meta value to retrieve your client from the provider configure method
		client := meta.(*apiClient)

		directoryService, diags := client.NewDirectoryService()
		if diags.HasError() {
			return diags
		}

		groupsService, diags := GetGroupsService(directoryService)
		if diags.HasError() {
			return diags
		}

		group, err := groupsService.Get(d.Get("email").(string)).Do()
		if err != nil {
			return diag.FromErr(err)
		}

		if group == nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("No group was returned for %s.", d.Get("email").(string)),
			})

			return diags
		}

		d.SetId(group.Id)
	}

	return resourceGroupRead(ctx, d, meta)
}
