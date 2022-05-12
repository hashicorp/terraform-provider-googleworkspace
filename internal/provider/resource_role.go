package googleworkspace

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-googleworkspace-pf/internal/model"
	"log"
)

type resourceRoleType struct{}

// GetSchema Role Resource
func (r resourceRoleType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Role resource in the Terraform Googleworkspace provider. Role resides " +
			"under the `https://www.googleapis.com/auth/admin.directory.rolemanagement` client scope.",
		Attributes: map[string]tfsdk.Attribute{
			"name": {
				Description: "Name of the role.",
				Type:        types.StringType,
				Required:    true,
			},
			"description": {
				Description: "A short description of the role.",
				Type:        types.StringType,
				Optional:    true,
			},
			"privileges": {
				Description: "The set of privileges that are granted to this role.",
				Required:    true,
				Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
					"service_id": {
						Description: "The obfuscated ID of the service this privilege is for.",
						Required:    true,
						Type:        types.StringType,
					},
					"privilege_name": {
						Description: "The name of the privilege.",
						Required:    true,
						Type:        types.StringType,
					},
				}, tfsdk.SetNestedAttributesOptions{}),
			},
			"is_system_role": {
				Description: "Returns true if this is a pre-defined system role.",
				Type:        types.BoolType,
				Computed:    true,
			},
			"is_super_admin_role": {
				Description: "Returns true if the role is a super admin role.",
				Type:        types.BoolType,
				Computed:    true,
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

type roleResource struct {
	provider provider
}

func (r resourceRoleType) NewResource(_ context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)

	return roleResource{
		provider: p,
	}, diags
}

// Create a new Role
func (r roleResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from "+
				"another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan model.RoleResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleReq := RolePlanToObj(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Creating Role %s", plan.Name.Value)
	rolesService := GetRolesService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	roleObj, err := rolesService.Insert(r.provider.customer, &roleReq).Do()
	if err != nil {
		resp.Diagnostics.AddError("error while trying to create role", err.Error())
		return
	}

	if roleObj == nil {
		resp.Diagnostics.AddError(fmt.Sprintf("no role was returned for %s", plan.Name.Value),
			"object returned was nil")
		return
	}

	role := SetRoleData(&plan, roleObj)

	diags = resp.State.Set(ctx, role)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Finished creating Role %s: %s", role.ID.Value, role.Name.Value)
}

// Read Role information
func (r roleResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state model.RoleResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	role := GetRoleData(&r.provider, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if role.ID.Null {
		resp.State.RemoveResource(ctx)
		log.Printf("[DEBUG] Removed Role from state because it was not found %s", state.ID.Value)
		return
	}

	diags = resp.State.Set(ctx, role)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	log.Printf("[DEBUG] Finished getting Role %s: %s", state.ID.Value, role.Name.Value)
}

// Update Role
func (r roleResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Retrieve values from plan
	var plan model.RoleResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state model.RoleResourceData
	diags = req.State.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleReq := RolePlanToObj(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Updating Role %s: %s", state.ID.Value, plan.Name.Value)
	rolesService := GetRolesService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	roleObj, err := rolesService.Update(r.provider.customer, state.ID.Value, &roleReq).Do()
	if err != nil {
		resp.Diagnostics.AddError("error while trying to update role", err.Error())
		return
	}

	if roleObj == nil {
		resp.Diagnostics.AddError(fmt.Sprintf("no role was returned for %s", plan.Name.Value),
			"object returned was nil")
		return
	}

	role := SetRoleData(&plan, roleObj)

	diags = resp.State.Set(ctx, role)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Finished updating Role %q: %#v", role.ID.Value, role.Name.Value)
}

// Delete Role
func (r roleResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state model.RoleResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Deleting Rolet %s: %s", state.ID.Value, state.Name.Value)
	rolesService := GetRolesService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := rolesService.Delete(r.provider.customer, state.ID.Value).Do()
	if err != nil {
		state.ID = types.String{Value: handleNotFoundError(err, state.ID.Value, &resp.Diagnostics)}
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.State.RemoveResource(ctx)
	log.Printf("[DEBUG] Finished deleting Role %s: %s", state.ID.Value, state.Name.Value)
}

// ImportState Role
func (r roleResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

//func getRole(d *schema.ResourceData) *directory.Role {
//	role := &directory.Role{
//		RoleName:        d.Get("name").(string),
//		RoleDescription: d.Get("description").(string),
//	}
//
//	privileges := d.Get("privileges").(*schema.Set)
//	for _, pMap := range privileges.List() {
//		priv := pMap.(map[string]interface{})
//		role.RolePrivileges = append(role.RolePrivileges, &directory.RoleRolePrivileges{
//			PrivilegeName: priv["privilege_name"].(string),
//			ServiceId:     priv["service_id"].(string),
//		})
//	}
//
//	if d.Id() != "" {
//		id, _ := strconv.ParseInt(d.Id(), 10, 64)
//		role.RoleId = id
//	}
//	return role
//}
//
//func setRole(d *schema.ResourceData, role *directory.Role) diag.Diagnostics {
//	var diags diag.Diagnostics
//
//	d.SetId(strconv.FormatInt(role.RoleId, 10))
//	d.Set("name", role.RoleName)
//	d.Set("description", role.RoleDescription)
//	d.Set("is_system_role", role.IsSystemRole)
//	d.Set("is_super_admin_role", role.IsSuperAdminRole)
//	d.Set("etag", role.Etag)
//
//	privileges := make([]interface{}, len(role.RolePrivileges))
//	for i, priv := range role.RolePrivileges {
//		privileges[i] = map[string]interface{}{
//			"service_id":     priv.ServiceId,
//			"privilege_name": priv.PrivilegeName,
//		}
//	}
//	if err := d.Set("privileges", privileges); err != nil {
//		diags = append(diags, diag.Diagnostic{
//			Severity:      diag.Error,
//			Summary:       "Error setting attribute",
//			Detail:        err.Error(),
//			AttributePath: cty.IndexStringPath("privileges"),
//		})
//	}
//
//	return diags
//}
