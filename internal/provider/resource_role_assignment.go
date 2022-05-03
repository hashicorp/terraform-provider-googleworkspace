package googleworkspace

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	directory "google.golang.org/api/admin/directory/v1"
	"log"
	"strconv"
)

type resourceRoleAssignmentType struct{}

// GetSchema Role Assignment Resource
func (r resourceRoleAssignmentType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Role Assignment resource in the Terraform Googleworkspace provider. Role Assignment resides " +
			"under the `https://www.googleapis.com/auth/admin.directory.rolemanagement` client scope.",
		Attributes: map[string]tfsdk.Attribute{
			"role_id": {
				Description: "The ID of the role that is assigned.",
				Type:        types.StringType,
				Required:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
			},
			"assigned_to": {
				Description: "The unique ID of the user this role is assigned to.",
				Type:        types.StringType,
				Required:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
			},
			"scope_type": {
				Description: "The scope in which this role is assigned. Valid values are 'CUSTOMER' or 'ORG_UNIT'",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []tfsdk.AttributeValidator{
					stringInSliceValidator{
						stringOptions: []string{"CUSTOMER", "ORG_UNIT"},
					},
				},
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						ValType:    types.StringType,
						DefaultVal: "CUSTOMER",
					},
					tfsdk.RequiresReplace(),
				},
			},
			"org_unit_id": {
				Description: "If the role is restricted to an organization unit, this contains the ID for the organization unit the exercise of this role is restricted to.",
				Type:        types.StringType,
				Optional:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
				//DiffSuppressFunc: diffSuppressOrgUnitId,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "Role Assignment identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

type roleAssignmentResourceData struct {
	ID         types.String `tfsdk:"id"`
	RoleId     types.String `tfsdk:"role_id"`
	AssignedTo types.String `tfsdk:"assigned_to"`
	ScopeType  types.String `tfsdk:"scope_type"`
	OrgUnitId  types.String `tfsdk:"org_unit_id"`
}

type roleAssignmentResource struct {
	provider provider
}

func (r resourceRoleAssignmentType) NewResource(_ context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)

	return roleAssignmentResource{
		provider: p,
	}, diags
}

// Create a new role assignment
func (r roleAssignmentResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from "+
				"another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan roleAssignmentResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleAssignmentReq := RoleAssignmentPlanToObj(&plan, &resp.Diagnostics)

	log.Printf("[DEBUG] Creating Role Assignment %q: %#v", plan.ID.Value, plan.ID.Value)
	roleAssignmentsService := GetRoleAssignmentsService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	roleAssignmentObj, err := roleAssignmentsService.Insert(r.provider.customer, &roleAssignmentReq).Do()
	if err != nil {
		resp.Diagnostics.AddError("error while trying to create role assignment", err.Error())
		return
	}

	if roleAssignmentObj == nil {
		resp.Diagnostics.AddError(fmt.Sprintf("no role assignment was returned for %s:%s", plan.RoleId.Value, plan.AssignedTo.Value), "object returned was nil")
		return
	}

	roleAssignment := SetRoleAssignmentData(roleAssignmentObj)

	diags = resp.State.Set(ctx, roleAssignment)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Finished creating role assignment %s: %s", roleAssignment.ID.Value, roleAssignment.RoleId.Value)
}

// Read role assignment information
func (r roleAssignmentResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state roleAssignmentResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleAssignment := GetRoleAssignmentData(&r.provider, state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if roleAssignment.ID.Null {
		resp.State.RemoveResource(ctx)
		log.Printf("[DEBUG] Remove Role Assignment from state because it was not found %s", state.ID.Value)
		return
	}

	diags = resp.State.Set(ctx, roleAssignment)
	if resp.Diagnostics.HasError() {
		return
	}
	log.Printf("[DEBUG] Finished getting Role Assignment %s: %s", state.ID.Value, roleAssignment.RoleId.Value)
}

// Update resource
func (r roleAssignmentResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
}

// Delete role assignment
func (r roleAssignmentResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state roleAssignmentResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Deleting Role Assignment %q: %#v", state.ID.Value, state.ID.Value)
	roleAssignmentsService := GetRoleAssignmentsService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := roleAssignmentsService.Delete(r.provider.customer, state.ID.Value).Do()
	if err != nil {
		state.ID = types.String{Value: handleNotFoundError(err, state.ID.Value, &resp.Diagnostics)}
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.State.RemoveResource(ctx)
	log.Printf("[DEBUG] Finished deleting Role Assignment %s: %s", state.ID.Value, state.ID.Value)
}

// ImportState role assignment
func (r roleAssignmentResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

// GetRoleAssignmentData from common Role Assignment object
func GetRoleAssignmentData(prov *provider, roleAssignmentData roleAssignmentResourceData, diags *diag.Diagnostics) roleAssignmentResourceData {
	roleAssignmentsService := GetRoleAssignmentsService(prov, diags)
	if diags.HasError() {
		return roleAssignmentResourceData{}
	}

	log.Printf("[DEBUG] Getting Role Assignment %q: %#v", roleAssignmentData.ID.Value, roleAssignmentData.ID.Value)

	roleAssignmentObj, err := roleAssignmentsService.Get(prov.customer, roleAssignmentData.ID.Value).Do()
	if err != nil {
		roleAssignmentData.ID = types.String{Value: handleNotFoundError(err, roleAssignmentData.ID.Value, diags)}
		if diags.HasError() {
			return roleAssignmentResourceData{}
		}
	}

	if roleAssignmentObj == nil {
		diags.AddError(fmt.Sprintf("no role assignment was returned for %s", roleAssignmentData.RoleId.Value), "object returned was nil")
		return roleAssignmentResourceData{}
	}

	return SetRoleAssignmentData(roleAssignmentObj)
}

// SetRoleAssignmentData from common Role Assignment object
func SetRoleAssignmentData(roleAssignmentObj *directory.RoleAssignment) roleAssignmentResourceData {
	return roleAssignmentResourceData{
		ID:         types.String{Value: strconv.FormatInt(roleAssignmentObj.RoleAssignmentId, 10)},
		RoleId:     types.String{Value: strconv.FormatInt(roleAssignmentObj.RoleId, 10)},
		AssignedTo: types.String{Value: roleAssignmentObj.AssignedTo},
		ScopeType:  types.String{Value: roleAssignmentObj.ScopeType},
		OrgUnitId:  types.String{Value: roleAssignmentObj.OrgUnitId, Null: roleAssignmentObj.OrgUnitId == ""},
	}
}
