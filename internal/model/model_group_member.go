package model

import "github.com/hashicorp/terraform-plugin-framework/types"

type GroupMemberResourceData struct {
	ID               types.String `tfsdk:"id"`
	GroupId          types.String `tfsdk:"group_id"`
	Email            types.String `tfsdk:"email"`
	Role             types.String `tfsdk:"role"`
	Type             types.String `tfsdk:"type"`
	Status           types.String `tfsdk:"status"`
	DeliverySettings types.String `tfsdk:"delivery_settings"`
	MemberId         types.String `tfsdk:"member_id"`
}
