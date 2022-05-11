package googleworkspace

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-googleworkspace-pf/internal/model"
	"google.golang.org/api/googleapi"
	"log"
	"strings"
)

type resourceGroupMemberType struct{}

// GetSchema Group Member Resource
func (r resourceGroupMemberType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Group Member resource manages Google Workspace Groups Members. Group Member resides under the " +
			"`https://www.googleapis.com/auth/admin.directory.group` client scope.",
		Attributes: map[string]tfsdk.Attribute{
			"group_id": {
				Description: "Identifies the group in the API request. The value can be the group's email address, " +
					"group alias, or the unique group ID.",
				Type:     types.StringType,
				Required: true,
			},
			"email": {
				Description: "The member's email address. A member can be a user or another group. This property is " +
					"required when adding a member to a group. The email must be unique and cannot be an alias of " +
					"another group. If the email address is changed, the API automatically reflects the email address changes.",
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
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
			"status": {
				Description: "Status of member.",
				Type:        types.StringType,
				Computed:    true,
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
						DefaultValue: types.String{Value: "ALL_MAIL"},
					},
				},
				Validators: []tfsdk.AttributeValidator{
					StringInSliceValidator{
						Options: []string{"ALL_MAIL", "DAILY", "DIGEST", "DISABLED", "NONE"},
					},
				},
			},
			"member_id": {
				Description: "The unique ID of the group member. A member id can be used as a member request URI's memberKey.",
				Type:        types.StringType,
				Computed:    true,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "Group Member identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

type groupMemberResource struct {
	provider provider
}

func (r resourceGroupMemberType) NewResource(_ context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)

	return groupMemberResource{
		provider: p,
	}, diags
}

// Create a new group member
func (r groupMemberResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from "+
				"another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan model.GroupMemberResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupMemberReq := GroupMemberPlanToObj(&plan)

	log.Printf("[DEBUG] Creating Group Member in group %s: %s", plan.GroupId.Value, plan.Email.Value)
	membersService := GetMembersService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	groupMemberObj, err := membersService.Insert(plan.GroupId.Value, &groupMemberReq).Do()
	if err != nil {
		resp.Diagnostics.AddError("error while trying to create group member", err.Error())
		return
	}

	if groupMemberObj == nil {
		resp.Diagnostics.AddError(fmt.Sprintf("no group member was returned for %s in %s",
			plan.Email.Value, plan.GroupId.Value), "object returned was nil")
		return
	}
	memberId := groupMemberObj.Id
	numInserts := 1

	// INSERT will respond with the Domain that will be created, after INSERT, the etag is updated along with the Domain,
	// once we get a consistent etag, we can feel confident that our Domain is also consistent
	cc := consistencyCheck{
		resourceType: "group member",
		timeout:      CreateTimeout,
	}
	err = retryTimeDuration(ctx, CreateTimeout, func() error {
		if cc.reachedConsistency(numInserts) {
			return nil
		}

		newDomain, retryErr := membersService.Get(r.provider.customer, memberId).IfNoneMatch(cc.lastEtag).Do()
		if googleapi.IsNotModified(retryErr) {
			cc.currConsistent += 1
		} else if retryErr != nil {
			return cc.is404(retryErr)
		} else {
			cc.handleNewEtag(newDomain.Etag)
		}

		return fmt.Errorf("timed out while waiting for %s to be inserted", cc.resourceType)
	})
	if err != nil {
		return
	}

	plan.ID.Value = memberId
	groupMember := GetGroupMemberData(&r.provider, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, groupMember)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Finished creating Group Member %s: %s", groupMember.ID.Value, groupMember.Email.Value)
}

// Read a group member
func (r groupMemberResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from "+
				"another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from state
	var state model.GroupMemberResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := GetGroupMemberData(&r.provider, &state, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &domain)
	resp.Diagnostics.Append(diags...)
}

// Update a group member
func (r groupMemberResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Retrieve values from plan
	var plan model.GroupMemberResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from plan
	var state model.GroupMemberResourceData
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupMemberReq := GroupMemberPlanToObj(&plan)

	log.Printf("[DEBUG] Updating Group Member %s: %s", plan.ID.Value, plan.Email.Value)
	membersService := GetMembersService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	groupMemberObj, err := membersService.Update(state.GroupId.Value, state.MemberId.Value, &groupMemberReq).Do()

	if err != nil {
		diags.AddError("error while trying to update group member", err.Error())
	}

	if groupMemberObj == nil {
		diags.AddError(fmt.Sprintf("no group member was returned for %s in group %s",
			plan.MemberId.Value, plan.GroupId.Value), "object returned was nil")
		return
	}

	numInserts := 1

	// UPDATE will respond with the Org Unit that will be created, however, it is eventually consistent
	// After UPDATE, the etag is updated along with the Org Unit,
	// once we get a consistent etag, we can feel confident that our Org Unit is also consistent
	cc := consistencyCheck{
		resourceType: "group member",
		timeout:      UpdateTimeout,
	}
	err = retryTimeDuration(ctx, UpdateTimeout, func() error {
		if cc.reachedConsistency(numInserts) {
			return nil
		}

		newOU, retryErr := membersService.Get(state.GroupId.Value, state.MemberId.Value).IfNoneMatch(cc.lastEtag).Do()
		if googleapi.IsNotModified(retryErr) {
			cc.currConsistent += 1
		} else if retryErr != nil {
			return cc.is404(retryErr)
		} else {
			cc.handleNewEtag(newOU.Etag)
		}

		return fmt.Errorf("timed out while waiting for %s to be updated", cc.resourceType)
	})
	if err != nil {
		return
	}

	groupMember := GetGroupMemberData(&r.provider, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, groupMember)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Finished updating Group Member %s: %s", state.ID.Value, plan.Email.Value)
}

// Delete a group member
func (r groupMemberResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state model.GroupMemberResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Removing Group Member %s", state.ID.Value)

	resp.State.RemoveResource(ctx)
}

// ImportState a group member
func (r groupMemberResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	parts := strings.Split(req.ID, "/")

	// id is of format "groups/<group_id>/members/<member_id>"
	if len(parts) != 4 {
		resp.Diagnostics.AddError("import id is not of the correct format",
			fmt.Sprintf("Group Member Id (%s) is not of the correct format (groups/<group_id>/members/<member_id>)", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("group_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("member_id"), parts[3])...)
}
