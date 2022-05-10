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
	"google.golang.org/api/googleapi"
	"log"
	"reflect"
)

type resourceGroupType struct{}

// GetSchema Group Resource
func (r resourceGroupType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Group resource manages Google Workspace Groups. Group resides under the " +
			"`https://www.googleapis.com/auth/admin.directory.group` client scope.",
		Attributes: map[string]tfsdk.Attribute{
			"email": {
				Description: "The group's email address. If your account has multiple domains," +
					"select the appropriate domain for the email address. The email must be unique.",
				Type:     types.StringType,
				Required: true,
			},
			"name": {
				Description: "The group's display name.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Description: "An extended description to help users determine the purpose of a group." +
					"For example, you can include information about who should join the group," +
					"the types of messages to send to the group, links to FAQs about the group, or related groups.",
				Type:     types.StringType,
				Optional: true,
				//ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(0, 4096)),
			},
			"admin_created": {
				Description: "Value is true if this group was created by an administrator rather than a user.",
				Type:        types.BoolType,
				Computed:    true,
			},
			"direct_members_count": {
				Description: "The number of users that are direct members of the group." +
					"If a group is a member (child) of this group (the parent)," +
					"members of the child group are not counted in the directMembersCount property of the parent group.",
				Type:     types.Int64Type,
				Computed: true,
			},
			"aliases": {
				Description: "asps.list of group's email addresses.",
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
			"non_editable_aliases": {
				Description: "asps.list of the group's non-editable alias email addresses that are outside of the " +
					"account's primary domain or subdomains. These are functioning email addresses used by the group.",
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Computed: true,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "Group identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

type groupResource struct {
	provider provider
}

func (r resourceGroupType) NewResource(_ context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)

	return groupResource{
		provider: p,
	}, diags
}

// Create a new group
func (r groupResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from "+
				"another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan model.GroupResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupReq := GroupPlanToObj(&plan)

	log.Printf("[DEBUG] Creating Group %s", plan.Email.Value)
	groupsService := GetGroupsService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	groupObj, err := groupsService.Insert(&groupReq).Do()
	if err != nil {
		resp.Diagnostics.AddError("error while trying to create group", err.Error())
		return
	}

	if groupObj == nil {
		resp.Diagnostics.AddError(fmt.Sprintf("no group was returned for %s", plan.Email.Value), "object returned was nil")
		return
	}
	numInserts := 1

	aliasesService := GetGroupAliasService(&r.provider, &resp.Diagnostics)
	// Insert all new aliases that weren't previously in state
	for _, alias := range plan.Aliases.Elems {
		aliasObj := directory.Alias{
			Alias: alias.(types.String).Value,
		}

		_, err := aliasesService.Insert(plan.ID.Value, &aliasObj).Do()
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("error while trying to add alias (%s) to group (%s)",
					alias.(types.String).Value, plan.Email.Value), err.Error())
		}
		numInserts += 1
	}

	// INSERT will respond with the Group that will be created, after INSERT, the etag is updated along with the Group,
	// once we get a consistent etag, we can feel confident that our Group is also consistent
	cc := consistencyCheck{
		resourceType: "group",
		timeout:      CreateTimeout,
	}
	err = retryTimeDuration(ctx, CreateTimeout, func() error {
		if cc.reachedConsistency(numInserts) {
			return nil
		}

		newGroup, retryErr := groupsService.Get(groupObj.Email).IfNoneMatch(cc.lastEtag).Do()
		if googleapi.IsNotModified(retryErr) {
			cc.currConsistent += 1
		} else if retryErr != nil {
			return cc.is404(retryErr)
		} else {
			cc.handleNewEtag(newGroup.Etag)
		}

		return fmt.Errorf("timed out while waiting for %s to be inserted", cc.resourceType)
	})
	if err != nil {
		return
	}

	plan.ID.Value = groupObj.Email
	group := GetGroupData(&r.provider, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, group)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Finished creating Group %s: %s", group.ID.Value, group.Email.Value)
}

// Read a group
func (r groupResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from "+
				"another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from state
	var state model.GroupResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	group := GetGroupData(&r.provider, &state, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &group)
	resp.Diagnostics.Append(diags...)
}

// Update a group
func (r groupResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Retrieve values from plan
	var plan model.GroupResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state model.GroupResourceData
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupReq := GroupPlanToObj(&plan)

	log.Printf("[DEBUG] Updating Org Unit %q: %#v", plan.ID.Value, plan.Name.Value)
	groupsService := GetGroupsService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	numInserts := 0
	if !reflect.DeepEqual(plan.Aliases, state.Aliases) {
		var stateAliases []string
		for _, sa := range typeListToSliceStrings(state.Aliases.Elems) {
			stateAliases = append(stateAliases, sa)
		}

		var planAliases []string
		for _, pa := range typeListToSliceStrings(plan.Aliases.Elems) {
			planAliases = append(planAliases, pa)
		}

		aliasesService := GetUserAliasService(&r.provider, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		// Remove old aliases that aren't in the new aliases list
		for _, alias := range stateAliases {
			if stringInSlice(planAliases, alias) {
				continue
			}

			err := aliasesService.Delete(state.ID.Value, alias).Do()
			if err != nil {
				resp.Diagnostics.AddError(fmt.Sprintf("error deleting alias (%s) from group (%s)", alias, state.Email.Value), err.Error())
				return
			}
		}

		// Insert all new aliases that weren't previously in state
		for _, alias := range planAliases {
			if stringInSlice(stateAliases, alias) {
				continue
			}

			aliasObj := directory.Alias{
				Alias: alias,
			}

			_, err := aliasesService.Insert(state.ID.Value, &aliasObj).Do()
			if err != nil {
				resp.Diagnostics.AddError(fmt.Sprintf("error inserting alias (%s) into group (%s)", alias, state.Email.Value), err.Error())
				return
			}
			numInserts += 1
		}
	}

	groupObj, err := groupsService.Update(state.ID.Value, &groupReq).Do()

	if err != nil {
		diags.AddError("error while trying to update org unit", err.Error())
	}

	if groupObj == nil {
		diags.AddError(fmt.Sprintf("no group was returned for %s", plan.Email.Value), "object returned was nil")
		return
	}

	numInserts += 1

	// UPDATE will respond with the Group that will be created, however, it is eventually consistent
	// After UPDATE, the etag is updated along with the Group (and any aliases),
	// once we get a consistent etag, we can feel confident that our User is also consistent
	cc := consistencyCheck{
		resourceType: "group",
		timeout:      UpdateTimeout,
	}
	err = retryTimeDuration(ctx, UpdateTimeout, func() error {
		var retryErr error

		if cc.reachedConsistency(numInserts) {
			return nil
		}

		newUser, retryErr := groupsService.Get(state.ID.Value).IfNoneMatch(cc.lastEtag).Do()
		if googleapi.IsNotModified(retryErr) {
			cc.currConsistent += 1
		} else if retryErr != nil {
			return fmt.Errorf("unexpected error during retries of %s: %s", cc.resourceType, retryErr)
		} else {
			cc.handleNewEtag(newUser.Etag)
		}

		return fmt.Errorf("timed out while waiting for %s to be updated", cc.resourceType)
	})
	if err != nil {
		resp.Diagnostics.AddError("error while trying to update user", err.Error())
		return
	}

	group := GetGroupData(&r.provider, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, group)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Finished updating Group %q: %#v", state.ID.Value, plan.Email.Value)
}

// Delete a group
func (r groupResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state model.GroupResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Removing Group %s: %s", state.ID.Value, state.Email.Value)

	resp.State.RemoveResource(ctx)
}

// ImportState a group
func (r groupResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
