package model

import "github.com/hashicorp/terraform-plugin-framework/types"

type GroupMembersResourceData struct {
	ID                       types.String `tfsdk:"id"`
	GroupId                  types.String `tfsdk:"group_id"`
	Members                  types.Set    `tfsdk:"members"`
	IncludeDerivedMembership types.Bool   `tfsdk:"include_derived_membership"`
}

type GroupMembersResourceMember struct {
	Email            types.String `tfsdk:"email"`
	Role             types.String `tfsdk:"role"`
	Type             types.String `tfsdk:"type"`
	Status           types.String `tfsdk:"status"`
	DeliverySettings types.String `tfsdk:"delivery_settings"`
	Id               types.String `tfsdk:"id"`
}
