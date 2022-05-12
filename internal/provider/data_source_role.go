package googleworkspace

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-provider-googleworkspace-pf/internal/model"
	directory "google.golang.org/api/admin/directory/v1"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type datasourceRoleType struct{}

func (t datasourceRoleType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	resourceType := resourceRoleType{}

	resourceSchema, diags := resourceType.GetSchema(ctx)
	if diags.HasError() {
		return tfsdk.Schema{}, diags
	}

	attrs := datasourceSchemaFromResourceSchema(resourceSchema.Attributes)
	addRequiredFieldsToSchema(attrs, "name")

	return tfsdk.Schema{
		Description: "Role data source in the Terraform Googleworkspace provider. Role resides " +
			"under the `https://www.googleapis.com/auth/admin.directory.rolemanagement` client scope.",
		Attributes: attrs,
	}, nil
}

type roleDatasource struct {
	provider provider
}

func (t datasourceRoleType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	p, diags := convertProviderType(in)

	return roleDatasource{
		provider: p,
	}, diags
}

func (d roleDatasource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data model.DomainResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := GetDomainData(&d.provider, &data, &resp.Diagnostics)
	if domain.ID.Value == "" {
		resp.Diagnostics.AddError("object does not exist",
			fmt.Sprintf("Domain %s does not exist", data.DomainName.Value))
	}

	diags = resp.State.Set(ctx, domain)
	resp.Diagnostics.Append(diags...)
}

func GetRoleData(prov *provider, plan *model.RoleResourceData, diags *diag.Diagnostics) *model.RoleResourceData {
	rolesService := GetRolesService(prov, diags)
	log.Printf("[DEBUG] Getting Role %s", plan.Name.Value)

	roleObj, err := rolesService.Get(prov.customer, plan.Name.Value).Do()
	if err != nil {
		plan.ID.Value = handleNotFoundError(err, plan.ID.Value, diags)
	}

	if roleObj == nil {
		diags.AddError("returned obj is nil", fmt.Sprintf("GET %s returned nil object",
			plan.Name.Value))
	}

	return SetRoleData(plan, roleObj)
}

func SetRoleData(plan *model.RoleResourceData, obj *directory.Role) *model.RoleResourceData {
	var privileges types.Set
	for _, priv := range obj.RolePrivileges {
		mem := model.RoleResourcePrivilege{
			ServiceId: types.String{
				Value: priv.ServiceId,
			},
			PrivilegeName: types.String{
				Value: priv.PrivilegeName,
			},
		}

		privileges.Elems = append(privileges.Elems, mem)
	}

	return &model.RoleResourceData{
		ID: types.String{
			Value: strconv.FormatInt(obj.RoleId, 10),
		},
		Name: types.String{
			Value: obj.RoleName,
		},
		Description: types.String{
			Value: obj.RoleDescription,
			Null:  plan.Description.Null,
		},
		Privileges: privileges,
		IsSystemRole: types.Bool{
			Value: obj.IsSystemRole,
		},
		IsSuperAdminRole: types.Bool{
			Value: obj.IsSuperAdminRole,
		},
	}
}

//func dataSourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	client := meta.(*apiClient)
//
//	directoryService, diags := client.NewDirectoryService()
//	if diags.HasError() {
//		return diags
//	}
//
//	rolesService, diags := GetRolesService(directoryService)
//	if diags.HasError() {
//		return diags
//	}
//
//	name := d.Get("name").(string)
//	var role *directory.Role
//	if err := rolesService.List(client.Customer).Pages(ctx, func(roles *directory.Roles) error {
//		for _, r := range roles.Items {
//			if r.RoleName == name {
//				role = r
//				return errors.New("role was found") // return error to stop pagination
//			}
//		}
//		return nil
//	}); role == nil && err != nil {
//		return diag.FromErr(err)
//	}
//
//	if role == nil {
//		return diag.Errorf("No role with name %q", name)
//	}
//
//	if diags := setRole(d, role); diags.HasError() {
//		return diags
//	}
//
//	return diags
//}
