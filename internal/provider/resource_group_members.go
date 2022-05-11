package googleworkspace

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-googleworkspace-pf/internal/model"
	directory "google.golang.org/api/admin/directory/v1"
	"log"
	"strings"
)

const deliverySettingsDefault = "ALL_MAIL"

type resourceGroupMembersType struct{}

// GetSchema Group Members Resource
func (r resourceGroupMembersType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Group Members resource manages Google Workspace Groups Members. Group Members resides under the " +
			"`https://www.googleapis.com/auth/admin.directory.group` client scope.",
		Attributes: map[string]tfsdk.Attribute{
			"group_id": {
				Description: "Identifies the group in the API request. The value can be the group's email address, " +
					"group alias, or the unique group ID.",
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
			},
			"members": {
				Description: "The members of the group",
				Optional:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"email": {
						Description: "The member's email address. A member can be a user or another group. This property is" +
							"required when adding a member to a group. The email must be unique and cannot be an alias of " +
							"another group. If the email address is changed, the API automatically reflects the email address changes.",
						Type:     types.StringType,
						Required: true,
					},
					"role": {
						Description: "The member's role in a group. The API returns an error for cycles in group memberships. " +
							"For example, if group1 is a member of group2, group2 cannot be a member of group1. " +
							"Acceptable values are: " +
							"`MANAGER`: This role is only available if the Google Groups for Business is " +
							"enabled using the Admin Console. A `MANAGER` role can do everything done by an `OWNER` role except " +
							"make a member an `OWNER` or delete the group. A group can have multiple `MANAGER` members. " +
							"`MEMBER`: This role can subscribe to a group, view discussion archives, and view the group's " +
							"membership list. " +
							"`OWNER`: This role can send messages to the group, add or remove members, change member roles, " +
							"change group's settings, and delete the group. An OWNER must be a member of the group. " +
							"A group can have more than one OWNER.",
						Type:     types.StringType,
						Optional: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							DefaultModifier{
								DefaultValue: types.String{Value: "MEMBER"},
							},
						},
						Validators: []tfsdk.AttributeValidator{
							StringInSliceValidator{
								Options: []string{"MANAGER", "MEMBER", "OWNER"},
							},
						},
					},
					"type": {
						Description: "The type of group member. Acceptable values are: " +
							"`CUSTOMER`: The member represents all users in a domain. An email address is not returned and the " +
							"ID returned is the customer ID. " +
							"`GROUP`: The member is another group. " +
							"`USER`: The member is a user.",
						Type:     types.StringType,
						Optional: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							DefaultModifier{
								DefaultValue: types.String{Value: "USER"},
							},
						},
						Validators: []tfsdk.AttributeValidator{
							StringInSliceValidator{
								Options: []string{"CUSTOMER", "GROUP", "USER"},
							},
						},
					},
					"delivery_settings": {
						Description: "Defines mail delivery preferences of member. Acceptable values are:" +
							"`ALL_MAIL`: All messages, delivered as soon as they arrive. " +
							"`DAILY`: No more than one message a day. " +
							"`DIGEST`: Up to 25 messages bundled into a single message. " +
							"`DISABLED`: Remove subscription. " +
							"`NONE`: No messages.",
						Type:     types.StringType,
						Optional: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							DefaultModifier{
								DefaultValue: types.String{Value: deliverySettingsDefault},
							},
						},
						Validators: []tfsdk.AttributeValidator{
							StringInSliceValidator{
								Options: []string{"ALL_MAIL", "DAILY", "DIGEST", "DISABLED", "NONE"},
							},
						},
					},
					"status": {
						Description: "Status of member.",
						Type:        types.StringType,
						Computed:    true,
					},
					"id": {
						Description: "The unique ID of the group member. A member id can be used as a member request URI's memberKey.",
						Type:        types.StringType,
						Computed:    true,
					},
				}, tfsdk.SetNestedAttributesOptions{}),
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "Group Members identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

type groupMembersResource struct {
	provider provider
}

func (r resourceGroupMembersType) NewResource(_ context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)

	return groupMembersResource{
		provider: p,
	}, diags
}

// Create new group members
func (r groupMembersResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from "+
				"another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan model.GroupMembersResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupMemberReqs := GroupMembersPlanToObj(ctx, &plan, &diags)

	membersService := GetMembersService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var memObjs []*directory.Member
	for _, memReq := range groupMemberReqs {
		log.Printf("[DEBUG] Creating Group Member in group %s: %s", plan.GroupId.Value, memReq.Email)

		mo, err := membersService.Insert(plan.GroupId.Value, &memReq).Do()
		if err != nil {
			resp.Diagnostics.AddError("error while trying to create group member", err.Error())
		}

		if mo == nil {
			resp.Diagnostics.AddError("object returned was nil", fmt.Sprintf("no group member was returned for %s", memReq.Email))
		}

		memObjs = append(memObjs, mo)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	groupMembers := SetGroupMembersData(&plan, memObjs)

	diags = resp.State.Set(ctx, groupMembers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Finished creating Group Members in group %s: %s", groupMembers.GroupId.Value, groupMembers.ID.Value)
}

// Read group members information
func (r groupMembersResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state model.GroupMembersResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupMembers := GetGroupMembersData(ctx, &r.provider, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if groupMembers.ID.Null {
		resp.State.RemoveResource(ctx)
		log.Printf("[DEBUG] Removed Org Unit from state because it was not found %s", state.ID.Value)
		return
	}

	diags = resp.State.Set(ctx, groupMembers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	log.Printf("[DEBUG] Finished getting Group Members in group %s: %s", groupMembers.GroupId.Value, groupMembers.ID.Value)
}

// Update group members
func (r groupMembersResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Retrieve values from plan
	var plan model.GroupMembersResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state model.GroupMembersResourceData
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupMemberReqs := GroupMembersPlanToObj(ctx, &plan, &diags)

	log.Printf("[DEBUG] Updating Group Members: %#v", state.ID.Value)
	membersService := GetMembersService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var planned []string

	for _, m := range plan.Members.Elems {
		var mem model.GroupMembersResourceMember
		d := m.(types.Object).As(ctx, &mem, types.ObjectAsOptions{})
		resp.Diagnostics.Append(d...)

		planned = append(planned, mem.Email.Value)
	}

	for _, m := range state.Members.Elems {
		var mem model.GroupMembersResourceMember
		d := m.(types.Object).As(ctx, &mem, types.ObjectAsOptions{})
		resp.Diagnostics.Append(d...)

		if !stringInSlice(planned, mem.Email.Value) {
			err := membersService.Delete(state.GroupId.Value, mem.Id.Value).Do()
			if err != nil {
				resp.Diagnostics.AddError("error deleting object on update", err.Error())
				if resp.Diagnostics.HasError() {
					return
				}
			}
		}
	}

	var memObjs []*directory.Member
	for _, memReq := range groupMemberReqs {
		var err error
		var mo *directory.Member

		if memReq.Id == "" {
			log.Printf("[DEBUG] Creating Group Member in group %s: %s", plan.GroupId.Value, memReq.Email)
			mo, err = membersService.Insert(plan.GroupId.Value, &memReq).Do()
		} else {
			log.Printf("[DEBUG] Updating Group Member in group %s: %s", plan.GroupId.Value, memReq.Email)
			mo, err = membersService.Update(plan.GroupId.Value, memReq.Id, &memReq).Do()
		}

		if err != nil {
			resp.Diagnostics.AddError("error while trying to update group members", err.Error())
		}

		if mo == nil {
			resp.Diagnostics.AddError("object returned was nil", fmt.Sprintf("no group member was returned for %s", memReq.Email))
		}

		memObjs = append(memObjs, mo)
	}

	groupMembers := SetGroupMembersData(&plan, memObjs)

	diags = resp.State.Set(ctx, groupMembers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Finished updating Group Members %s", state.ID.Value)
}

// Delete group members
func (r groupMembersResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state model.GroupMembersResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Deleting Group Members : %s", state.ID.Value)
	membersService := GetMembersService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, m := range state.Members.Elems {
		var mem model.GroupMembersResourceMember
		d := m.(types.Object).As(ctx, &mem, types.ObjectAsOptions{})
		resp.Diagnostics.Append(d...)

		err := membersService.Delete(state.GroupId.Value, mem.Id.Value).Do()
		if err != nil {
			state.ID = types.String{Value: handleNotFoundError(err, state.ID.Value, &resp.Diagnostics)}
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}

	resp.State.RemoveResource(ctx)
	log.Printf("[DEBUG] Finished deleting Group Members: %s", state.ID.Value)
}

// ImportState group members
func (r groupMembersResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	parts := strings.Split(req.ID, "/")

	// id is of format "groups/<group_id>"
	if len(parts) != 2 {
		resp.Diagnostics.AddError("import id is not of the correct format",
			fmt.Sprintf("Group Members Id (%s) is not of the correct format (groups/<group_id>)", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("group_id"), parts[1])...)
}
