// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	directory "google.golang.org/api/admin/directory/v1"
)

func dataSourceRole() *schema.Resource {
	rSchema := datasourceSchemaFromResourceSchema(resourceRole().Schema)
	addRequiredFieldsToSchema(rSchema, "name")

	return &schema.Resource{
		Description: "Role data source in the Terraform Googleworkspace provider. Role resides " +
			"under the `https://www.googleapis.com/auth/admin.directory.rolemanagement` client scope.",

		ReadContext: dataSourceRoleRead,

		Schema: rSchema,
	}
}

func dataSourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	rolesService, diags := GetRolesService(directoryService)
	if diags.HasError() {
		return diags
	}

	name := d.Get("name").(string)
	var role *directory.Role
	if err := rolesService.List(client.Customer).Pages(ctx, func(roles *directory.Roles) error {
		for _, r := range roles.Items {
			if r.RoleName == name {
				role = r
				return errors.New("role was found") // return error to stop pagination
			}
		}
		return nil
	}); role == nil && err != nil {
		return diag.FromErr(err)
	}

	if role == nil {
		return diag.Errorf("No role with name %q", name)
	}

	if diags := setRole(d, role); diags.HasError() {
		return diags
	}

	return diags
}
