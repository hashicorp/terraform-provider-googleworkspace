// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGroupMembers() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceGroupMembers().Schema)
	addRequiredFieldsToSchema(dsSchema, "group_id")
	dsSchema["include_derived_membership"] = &schema.Schema{
		Description: "If true, lists indirect group memberships",
		Type:        schema.TypeBool,
		Optional:    true,
	}

	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Group Members data source in the Terraform Googleworkspace provider. Group Members resides " +
			"under the `https://www.googleapis.com/auth/admin.directory.group` client scope.",

		ReadContext: dataSourceGroupMembersRead,

		Schema: dsSchema,
	}
}

func dataSourceGroupMembersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceGroupMembersRead(ctx, d, meta)
}
