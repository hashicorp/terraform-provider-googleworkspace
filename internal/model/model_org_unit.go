package model

import "github.com/hashicorp/terraform-plugin-framework/types"

type OrgUnitResourceData struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	BlockInheritance  types.Bool   `tfsdk:"block_inheritance"`
	OrgUnitId         types.String `tfsdk:"org_unit_id"`
	OrgUnitPath       types.String `tfsdk:"org_unit_path"`
	ParentOrgUnitId   types.String `tfsdk:"parent_org_unit_id"`
	ParentOrgUnitPath types.String `tfsdk:"parent_org_unit_path"`
}
