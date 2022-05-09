package model

import "github.com/hashicorp/terraform-plugin-framework/types"

type RoleAssignmentResourceData struct {
	ID         types.String `tfsdk:"id"`
	RoleId     types.Int64  `tfsdk:"role_id"`
	AssignedTo types.String `tfsdk:"assigned_to"`
	ScopeType  types.String `tfsdk:"scope_type"`
	OrgUnitId  types.String `tfsdk:"org_unit_id"`
}
