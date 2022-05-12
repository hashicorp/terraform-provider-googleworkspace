package model

import "github.com/hashicorp/terraform-plugin-framework/types"

type RoleResourceData struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	Privileges       types.Set    `tfsdk:"privileges"`
	IsSystemRole     types.Bool   `tfsdk:"is_system_role"`
	IsSuperAdminRole types.Bool   `tfsdk:"is_super_admin_role"`
}

type RoleResourcePrivilege struct {
	ServiceId     types.String `tfsdk:"service_id"`
	PrivilegeName types.String `tfsdk:"privilege_name"`
}
