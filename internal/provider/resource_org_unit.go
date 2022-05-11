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

// GetOrgUnitId returns the org unit id that is prefixed with "id:..."
// We save it un-prefixed in state since some resources don't want the prefix
// but we want it for org resource calls
func GetOrgUnitId(orgUnitId string) string {
	if !strings.HasPrefix(orgUnitId, "id:") {
		orgUnitId = "id:" + orgUnitId
	}

	return orgUnitId
}

type resourceOrgUnitType struct{}

// GetSchema OrgUnit Resource
func (r resourceOrgUnitType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "OrgUnit resource manages Google Workspace OrgUnits. Org Unit resides " +
			"under the `https://www.googleapis.com/auth/admin.directory.orgunit` client scope.",
		Attributes: map[string]tfsdk.Attribute{
			"name": {
				Description: "The organizational unit's path name. For example, an organizational unit's name within the " +
					"/corp/support/sales_support parent path is sales_support.",
				Type:     types.StringType,
				Required: true,
			},
			"description": {
				Description: "Description of the organizational unit.",
				Type:        types.StringType,
				Optional:    true,
			},
			"block_inheritance": {
				Description: "Determines if a sub-organizational unit can inherit the settings of the parent organization. " +
					"False means a sub-organizational unit inherits the settings of the nearest parent organizational unit. " +
					"For more information on inheritance and users in an organization structure, see the " +
					"[administration help center](https://support.google.com/a/answer/4352075).",
				Type:     types.BoolType,
				Optional: true,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.Bool{Value: false},
					},
				},
			},
			"org_unit_id": {
				Description: "The unique ID of the organizational unit.",
				Type:        types.StringType,
				Computed:    true,
			},
			"org_unit_path": {
				Description: "The full path to the organizational unit. The orgUnitPath is a derived property. " +
					"When listed, it is derived from parentOrgunitPath and organizational unit's name. For example, " +
					"for an organizational unit named 'apps' under parent organization '/engineering', the orgUnitPath " +
					"is '/engineering/apps'. In order to edit an orgUnitPath, either update the name of the organization " +
					"or the parentOrgunitPath. A user's organizational unit determines which Google Workspace services " +
					"the user has access to. If the user is moved to a new organization, the user's access changes. " +
					"For more information about organization structures, see the [administration help center](https://support.google.com/a/answer/4352075). " +
					"For more information about moving a user to a different organization, see " +
					"[chromeosdevices.update a user](https://developers.google.com/admin-sdk/directory/v1/guides/manage-users#update_user).",
				Type:     types.StringType,
				Computed: true,
			},
			"parent_org_unit_id": {
				Description: "The unique ID of the parent organizational unit.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []tfsdk.AttributeValidator{
					ExactlyOneOfValidator{
						RequiredAttrs: []string{"parent_org_unit_id", "parent_org_unit_path"},
					},
				},
			},
			"parent_org_unit_path": {
				Description: "The organizational unit's parent path. For example, /corp/sales is the parent path for " +
					"/corp/sales/sales_support organizational unit.",
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				Validators: []tfsdk.AttributeValidator{
					ExactlyOneOfValidator{
						RequiredAttrs: []string{"parent_org_unit_id", "parent_org_unit_path"},
					},
				},
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "Org Unit identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

type orgUnitResource struct {
	provider provider
}

func (r resourceOrgUnitType) NewResource(_ context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)

	return orgUnitResource{
		provider: p,
	}, diags
}

// Create a new org unit
func (r orgUnitResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from "+
				"another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan model.OrgUnitResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgUnitReq := OrgUnitPlanToObj(&plan)

	log.Printf("[DEBUG] Creating Org Unit %s", plan.Name.Value)
	orgUnitsService := GetOrgUnitsService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	orgUnitObj, err := orgUnitsService.Insert(r.provider.customer, &orgUnitReq).Do()
	if err != nil {
		resp.Diagnostics.AddError("error while trying to create org unit", err.Error())
		return
	}

	if orgUnitObj == nil {
		resp.Diagnostics.AddError(fmt.Sprintf("no org unit was returned for %s", plan.Name.Value), "object returned was nil")
		return
	}

	orgUnitId := strings.TrimPrefix(orgUnitObj.OrgUnitId, "id:")
	numInserts := 1

	// INSERT will respond with the Org Unit that will be created, however, it is eventually consistent
	// After INSERT, the etag is updated along with the Org Unit,
	// once we get a consistent etag, we can feel confident that our Org Unit is also consistent
	cc := consistencyCheck{
		resourceType: "org unit",
		timeout:      CreateTimeout,
	}
	err = retryTimeDuration(ctx, CreateTimeout, func() error {
		if cc.reachedConsistency(numInserts) {
			return nil
		}

		newOU, retryErr := orgUnitsService.Get(r.provider.customer, GetOrgUnitId(orgUnitId)).IfNoneMatch(cc.lastEtag).Do()
		if googleapi.IsNotModified(retryErr) {
			cc.currConsistent += 1
		} else if retryErr != nil {
			return cc.is404(retryErr)
		} else {
			cc.handleNewEtag(newOU.Etag)
		}

		return fmt.Errorf("timed out while waiting for %s to be inserted", cc.resourceType)
	})
	if err != nil {
		return
	}

	plan.ID.Value = orgUnitId
	orgUnit := GetOrgUnitData(&r.provider, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, orgUnit)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Finished creating Org Unit %s: %s", orgUnit.ID.Value, orgUnit.Name.Value)
}

// Read org unit information
func (r orgUnitResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state model.OrgUnitResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgUnit := GetOrgUnitData(&r.provider, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if orgUnit.ID.Null {
		resp.State.RemoveResource(ctx)
		log.Printf("[DEBUG] Removed Org Unit from state because it was not found %s", state.ID.Value)
		return
	}

	diags = resp.State.Set(ctx, orgUnit)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	log.Printf("[DEBUG] Finished getting Org Unit %s: %s", state.ID.Value, orgUnit.Name.Value)
}

// Update org unit
func (r orgUnitResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Retrieve values from plan
	var plan model.OrgUnitResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from plan
	var state model.OrgUnitResourceData
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgUnitReq := OrgUnitPlanToObj(&plan)

	log.Printf("[DEBUG] Updating Org Unit %q: %#v", plan.ID.Value, plan.Name.Value)
	orgUnitsService := GetOrgUnitsService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	orgUnitObj, err := orgUnitsService.Update(r.provider.customer, GetOrgUnitId(state.ID.Value), &orgUnitReq).Do()

	if err != nil {
		diags.AddError("error while trying to update org unit", err.Error())
	}

	if orgUnitObj == nil {
		diags.AddError(fmt.Sprintf("no org unit was returned for %s", plan.Name.Value), "object returned was nil")
		return
	}

	orgUnitId := strings.TrimPrefix(orgUnitObj.OrgUnitId, "id:")
	numInserts := 1

	// UPDATE will respond with the Org Unit that will be created, however, it is eventually consistent
	// After UPDATE, the etag is updated along with the Org Unit,
	// once we get a consistent etag, we can feel confident that our Org Unit is also consistent
	cc := consistencyCheck{
		resourceType: "org unit",
		timeout:      UpdateTimeout,
	}
	err = retryTimeDuration(ctx, UpdateTimeout, func() error {
		if cc.reachedConsistency(numInserts) {
			return nil
		}

		newOU, retryErr := orgUnitsService.Get(r.provider.customer, GetOrgUnitId(orgUnitId)).IfNoneMatch(cc.lastEtag).Do()
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

	plan.ID.Value = orgUnitId
	orgUnit := GetOrgUnitData(&r.provider, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, orgUnit)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Finished updating OrgUnit %q: %#v", state.ID.Value, plan.Name.Value)
}

// Delete org unit
func (r orgUnitResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state model.OrgUnitResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Deleting Org Unit %s: %s", state.ID.Value, state.Name.Value)
	orgUnitsService := GetOrgUnitsService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := orgUnitsService.Delete(r.provider.customer, GetOrgUnitId(state.ID.Value)).Do()
	if err != nil {
		state.ID = types.String{Value: handleNotFoundError(err, state.ID.Value, &resp.Diagnostics)}
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.State.RemoveResource(ctx)
	log.Printf("[DEBUG] Finished deleting Org Unit %s: %s", state.ID.Value, state.Name.Value)
}

// ImportState org unit
func (r orgUnitResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
