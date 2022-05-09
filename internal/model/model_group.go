package model

import "github.com/hashicorp/terraform-plugin-framework/types"

type GroupResourceData struct {
	ID                 types.String `tfsdk:"id"`
	Email              types.String `tfsdk:"email"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	AdminCreated       types.Bool   `tfsdk:"admin_created"`
	DirectMembersCount types.Int64  `tfsdk:"direct_members_count"`
	Aliases            types.List   `tfsdk:"aliases"`
	NonEditableAliases types.List   `tfsdk:"non_editable_aliases"`
}
