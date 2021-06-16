package googleworkspace

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	directory "google.golang.org/api/admin/directory/v1"
)

func dataSourcePrivileges() *schema.Resource {
	return &schema.Resource{
		Description: "Privileges data source in the Terraform Googleworkspace provider.",

		ReadContext: dataSourcePrivilegeRead,

		Schema: map[string]*schema.Schema{
			"etag": {
				Description: "ETag of the resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"items": {
				Description: "A list of Privilege resources.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_id": {
							Description: "The obfuscated ID of the service this privilege is for.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"etag": {
							Description: "ETag of the resource.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"is_org_unit_scopable": {
							Description: "If the privilege can be restricted to an organization unit.",
							Type:        schema.TypeBool,
							Computed:    true,
						},
						"privilege_name": {
							Description: "The name of the privilege.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"service_name": {
							Description: "The name of the service this privilege is for. Please note this field is empty for many privileges and may not be a reliable field to attempt to filter on",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourcePrivilegeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	privilegesService, diags := GetPrivilegesService(directoryService)
	if diags.HasError() {
		return diags
	}

	privileges, err := privilegesService.List(client.Customer).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(privileges.Etag)
	d.Set("etag", privileges.Etag)

	if err := d.Set("items", flattenAndPrunePrivileges(privileges.Items, make(map[string]bool))); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenAndPrunePrivileges(privileges []*directory.Privilege, duplicates map[string]bool) []interface{} {
	var result []interface{}
	for _, priv := range privileges {
		// these are the two fields passed to roles, so they are the fields
		// considered in uniqueness
		id := priv.PrivilegeName + ":" + priv.ServiceId
		if !duplicates[id] {
			result = append(result, map[string]interface{}{
				"service_id":           priv.ServiceId,
				"etag":                 priv.Etag,
				"is_org_unit_scopable": priv.IsOuScopable,
				"privilege_name":       priv.PrivilegeName,
				"service_name":         priv.ServiceName,
			})
			duplicates[id] = true
		}
		if len(priv.ChildPrivileges) > 0 {
			result = append(result, flattenAndPrunePrivileges(priv.ChildPrivileges, duplicates)...)
		}
	}
	return result
}
