package googleworkspace

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-provider-googleworkspace-pf/internal/model"
	directory "google.golang.org/api/admin/directory/v1"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type datasourceGroupMembersType struct{}

func (t datasourceGroupMembersType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	resourceType := resourceDomainType{}

	resourceSchema, diags := resourceType.GetSchema(ctx)
	if diags.HasError() {
		return tfsdk.Schema{}, diags
	}

	attrs := datasourceSchemaFromResourceSchema(resourceSchema.Attributes)
	addRequiredFieldsToSchema(attrs, "group_id")
	attrs["include_derived_membership"] = tfsdk.Attribute{
		Description: "If true, lists indirect group memberships. Defaults to false.",
		Type:        types.BoolType,
		Optional:    true,
		PlanModifiers: []tfsdk.AttributePlanModifier{
			DefaultModifier{
				DefaultValue: types.Bool{Value: false},
			},
		},
	}

	return tfsdk.Schema{
		Description: "Group Members data source in the Terraform Googleworkspace provider. Group Members resides " +
			"under the `https://www.googleapis.com/auth/admin.directory.group` client scope.",
		Attributes: attrs,
	}, nil
}

type groupMembersDatasource struct {
	provider provider
}

func (t datasourceGroupMembersType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	p, diags := convertProviderType(in)

	return groupMembersDatasource{
		provider: p,
	}, diags
}

func (d groupMembersDatasource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data model.GroupMembersResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupMembers := GetGroupMembersData(ctx, &d.provider, &data, &resp.Diagnostics)
	if groupMembers.ID.Value == "" {
		resp.Diagnostics.AddError("object does not exist",
			fmt.Sprintf("Group Members do not exist for group %s", data.GroupId.Value))
	}

	diags = resp.State.Set(ctx, groupMembers)
	resp.Diagnostics.Append(diags...)
}

func GetGroupMembersData(ctx context.Context, prov *provider, plan *model.GroupMembersResourceData, diags *diag.Diagnostics) *model.GroupMembersResourceData {
	membersService := GetMembersService(prov, diags)
	log.Printf("[DEBUG] Getting Group Members for group %s", plan.GroupId.Value)

	var groupMemberObjs []*directory.Member
	membersCall := membersService.List(plan.GroupId.Value).MaxResults(200).IncludeDerivedMembership(plan.IncludeDerivedMembership.Value)

	err := membersCall.Pages(ctx, func(resp *directory.Members) error {
		for _, member := range resp.Members {
			result = append(result, member)
		}

		return nil
	})
	if err != nil {
		plan.ID.Value = handleNotFoundError(err, plan.ID.Value, diags)
	}

	if groupMemberObjs == nil {
		diags.AddError("returned obj is nil", fmt.Sprintf("GET Group Members in %s returned nil object",
			plan.GroupId.Value))
	}

	return SetGroupMembersData(plan, groupMemberObjs)
}

func SetGroupMembersData(plan *model.GroupMembersResourceData, objs []*directory.Member) *model.GroupMembersResourceData {
	var members []model.GroupMembersResourceMember
	for _, mo := range objs {
		mem := model.GroupMembersResourceMember{
			Email:            types.String{Value: mo.Email},
			Role:             types.String{Value: mo.Role},
			Type:             types.String{Value: mo.Type},
			Status:           types.String{Value: mo.Status},
			DeliverySettings: types.String{Value: mo.DeliverySettings},
			Id:               types.String{Value: mo.Id},
		}

		members = append(members, mem)
	}

	return &model.GroupMembersResourceData{
		ID:                       types.String{Value: fmt.Sprintf("groups/%s", plan.GroupId)},
		GroupId:                  types.String{Value: plan.GroupId.Value},
		IncludeDerivedMembership: types.Bool{Value: plan.IncludeDerivedMembership.Value},
		Members:                  members,
	}
}
