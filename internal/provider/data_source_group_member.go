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

type datasourceGroupMemberType struct{}

func (t datasourceGroupMemberType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	resourceType := resourceDomainType{}

	resourceSchema, diags := resourceType.GetSchema(ctx)
	if diags.HasError() {
		return tfsdk.Schema{}, diags
	}

	attrs := datasourceSchemaFromResourceSchema(resourceSchema.Attributes)
	addRequiredFieldsToSchema(attrs, "group_id")
	addExactlyOneOfFieldsToSchema(attrs, "member_id", "email")

	return tfsdk.Schema{
		Description: "Group Member data source in the Terraform Googleworkspace provider. Group Member resides under the " +
			"`https://www.googleapis.com/auth/admin.directory.group` client scope.",
		Attributes: attrs,
	}, nil
}

type groupMemberDatasource struct {
	provider provider
}

func (t datasourceGroupMemberType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	p, diags := convertProviderType(in)

	return groupMemberDatasource{
		provider: p,
	}, diags
}

func (d groupMemberDatasource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data model.GroupMemberResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.MemberId.Null {
		data.MemberId = data.Email
	}

	groupMember := GetGroupMemberData(&d.provider, &data, &resp.Diagnostics)
	if groupMember.ID.Value == "" {
		resp.Diagnostics.AddError("object does not exist",
			fmt.Sprintf("Group Member %s does not exist in %s", data.MemberId.Value, data.GroupId.Value))
	}

	diags = resp.State.Set(ctx, groupMember)
	resp.Diagnostics.Append(diags...)
}

func GetGroupMemberData(prov *provider, plan *model.GroupMemberResourceData, diags *diag.Diagnostics) *model.GroupMemberResourceData {
	membersService := GetMembersService(prov, diags)
	log.Printf("[DEBUG] Getting Group Member %s", plan.MemberId.Value)

	groupMemberObj, err := membersService.Get(plan.GroupId.Value, plan.MemberId.Value).Do()
	if err != nil {
		plan.ID.Value = handleNotFoundError(err, plan.ID.Value, diags)
	}

	if groupMemberObj == nil {
		diags.AddError("returned obj is nil", fmt.Sprintf("GET %s in %s returned nil object",
			plan.MemberId.Value, plan.GroupId.Value))
	}

	return SetGroupMemberData(plan, groupMemberObj)
}

func SetGroupMemberData(plan *model.GroupMemberResourceData, obj *directory.Member) *model.GroupMemberResourceData {
	return &model.GroupMemberResourceData{
		ID:               types.String{Value: fmt.Sprintf("groups/%s/members/%s", plan.GroupId.Value, obj.Id)},
		GroupId:          types.String{Value: plan.GroupId.Value},
		Email:            types.String{Value: obj.Email},
		Role:             types.String{Value: obj.Role},
		Type:             types.String{Value: obj.Type},
		Status:           types.String{Value: obj.Status},
		DeliverySettings: types.String{Value: obj.DeliverySettings},
		MemberId:         types.String{Value: obj.Id},
	}
}
