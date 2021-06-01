package googleworkspace

import (
	"context"
	"errors"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	directory "google.golang.org/api/admin/directory/v1"
)

func dataSourceRole() *schema.Resource {
	return &schema.Resource{
		Description: "Roles data source in the Terraform Googleworkspace provider.",

		ReadContext: dataSourceRoleRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "ID of the role.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Name of the role.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "A short description of the role.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"privileges": {
				Description: "The set of privileges that are granted to this role.",
				Computed:    true,
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_id": {
							Description: "The obfuscated ID of the service this privilege is for.",
							Computed:    true,
							Type:        schema.TypeString,
						},
						"name": {
							Description: "The name of the privilege.",
							Computed:    true,
							Type:        schema.TypeString,
						},
					},
				},
			},
			"is_system_role": {
				Description: "Returns true if this is a pre-defined system role.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"is_super_admin_role": {
				Description: "Returns true if the role is a super admin role.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"kind": {
				Description: "The type of the API resource. This is always admin#directory#role.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"etag": {
				Description: "ETag of the resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
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

	d.SetId(strconv.FormatInt(role.RoleId, 10))
	d.Set("description", role.RoleDescription)
	d.Set("is_system_role", role.IsSystemRole)
	d.Set("is_super_admin_role", role.IsSuperAdminRole)
	d.Set("kind", role.Kind)
	d.Set("etag", role.Etag)

	privileges := make([]interface{}, len(role.RolePrivileges))
	for _, priv := range role.RolePrivileges {
		privileges = append(privileges, map[string]interface{}{
			"service_id": priv.ServiceId,
			"name":       priv.PrivilegeName,
		})
	}
	if err := d.Set("privileges", privileges); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
