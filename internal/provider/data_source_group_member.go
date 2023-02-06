// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGroupMember() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceGroupMember().Schema)
	addRequiredFieldsToSchema(dsSchema, "group_id")
	addExactlyOneOfFieldsToSchema(dsSchema, "member_id", "email")

	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Group Member data source in the Terraform Googleworkspace provider. Group Member resides under the " +
			"`https://www.googleapis.com/auth/admin.directory.group` client scope.",

		ReadContext: dataSourceGroupMemberRead,

		Schema: dsSchema,
	}
}

func dataSourceGroupMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Get("member_id") != "" {
		groupId := d.Get("group_id").(string)
		memberId := d.Get("member_id").(string)
		d.SetId(fmt.Sprintf("groups/%s/members/%s", groupId, memberId))
	} else {
		var diags diag.Diagnostics

		// use the meta value to retrieve your client from the provider configure method
		client := meta.(*apiClient)

		directoryService, diags := client.NewDirectoryService()
		if diags.HasError() {
			return diags
		}

		membersService, diags := GetMembersService(directoryService)
		if diags.HasError() {
			return diags
		}

		groupId := d.Get("group_id").(string)
		member, err := membersService.Get(groupId, d.Get("email").(string)).Do()
		if err != nil {
			return diag.FromErr(err)
		}

		if member == nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("No group member was returned for %s in group %s", d.Get("email").(string), groupId),
			})

			return diags
		}

		d.Set("member_id", member.Id)
		d.SetId(fmt.Sprintf("groups/%s/members/%s", groupId, member.Id))
	}

	return resourceGroupMemberRead(ctx, d, meta)
}
