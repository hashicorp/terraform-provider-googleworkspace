package googleworkspace

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	directory "google.golang.org/api/admin/directory/v1"
	"strconv"
)

func dataSourceRoleAssignments() *schema.Resource {
	return &schema.Resource{
		Description: "List all Role Assignments",

		ReadContext: dataSourceRoleAssignmentsRead,

		Schema: map[string]*schema.Schema{
			"items": {
				Description: "RoleAssignments",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role_id": {
							Description: "The ID of the role that is assigned.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"etag": {
							Description: "ETag of the resource.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"assigned_to": {
							Description: "The unique ID of the user this role is assigned to.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"scope_type": {
							Description: "The scope in which this role is assigned. Valid values are 'CUSTOMER' or 'ORG_UNIT'",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"org_unit_id": {
							Description: "If the role is restricted to an organization unit, this contains the ID for the organization unit the exercise of this role is restricted to.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"role_assignment_id": {
							Description: "ID of this roleAssignment.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceRoleAssignmentsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	roleAssignmentsService, diags := GetRoleAssignmentsService(directoryService)
	if diags.HasError() {
		return diags
	}

	pageToken := ""
	haveNextPage := true
	roleAssignments := make([]*directory.RoleAssignment, 0)

	for haveNextPage {
		page, err := roleAssignmentsService.List(client.Customer).PageToken(pageToken).Do()
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(page.Etag)
		for _, v := range page.Items {
			roleAssignments = append(roleAssignments, v)
		}
		pageToken = page.NextPageToken
		if pageToken == "" {
			haveNextPage = false
		}
	}

	var result []interface{}
	for _, v := range roleAssignments {
		result = append(result, map[string]interface{}{
			"role_id":            strconv.FormatInt(v.RoleId, 10),
			"etag":               v.Etag,
			"assigned_to":        v.AssignedTo,
			"scope_type":         v.ScopeType,
			"org_unit_id":        v.OrgUnitId,
			"role_assignment_id": strconv.FormatInt(v.RoleAssignmentId, 10),
		})
	}
	if err := d.Set("items", result); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
