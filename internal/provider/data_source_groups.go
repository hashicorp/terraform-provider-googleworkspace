package googleworkspace

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	directory "google.golang.org/api/admin/directory/v1"
)

func dataSourceGroups() *schema.Resource {
	// Generate datasource schema from resource
	dsGroupSchema := datasourceSchemaFromResourceSchema(resourceGroup().Schema)

	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Groups data source in the Terraform Googleworkspace provider. Groups resides " +
			"under the `https://www.googleapis.com/auth/admin.directory.group` client scope.",

		ReadContext: dataSourceGroupsRead,

		Schema: map[string]*schema.Schema{
			"groups": {
				Description: "A list of Group resources.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: dsGroupSchema,
				},
			},
		},
	}
}

func dataSourceGroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client := meta.(*apiClient)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	groupsService, diags := GetGroupsService(directoryService)
	if diags.HasError() {
		return diags
	}

	var result []*directory.Group
	err := groupsService.List().Customer(client.Customer).Pages(ctx, func(resp *directory.Groups) error {
		for _, group := range resp.Groups {
			result = append(result, group)
		}

		return nil
	})

	if err != nil {
		return handleNotFoundError(err, d, "groups")
	}

	if err := d.Set("groups", flattenGroups(result)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("groups")

	return diags
}

func flattenGroups(groups []*directory.Group) interface{} {
	var result []interface{}

	for _, group := range groups {
		result = append(result, flattenGroup(group))
	}

	return result
}

func flattenGroup(group *directory.Group) interface{} {

	result := map[string]interface{}{}
	result["admin_created"] = group.AdminCreated
	result["aliases"] = group.Aliases
	result["description"] = group.Description
	result["direct_members_count"] = group.DirectMembersCount
	result["email"] = group.Email
	result["etag"] = group.Etag
	result["id"] = group.Id
	result["name"] = group.Name
	result["non_editable_aliases"] = group.NonEditableAliases

	return result
}
