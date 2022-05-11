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

type datasourceGroupType struct{}

func (t datasourceGroupType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	resourceType := resourceGroupType{}

	resourceSchema, diags := resourceType.GetSchema(ctx)
	if diags.HasError() {
		return tfsdk.Schema{}, diags
	}

	attrs := datasourceSchemaFromResourceSchema(resourceSchema.Attributes)
	addExactlyOneOfFieldsToSchema(attrs, "id", "email")

	return tfsdk.Schema{
		Description: "Group data source in the Terraform Googleworkspace provider. Group resides under the " +
			"`https://www.googleapis.com/auth/admin.directory.group` client scope.",
		Attributes: attrs,
	}, nil
}

type groupDatasource struct {
	provider provider
}

func (t datasourceGroupType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	p, diags := convertProviderType(in)

	return groupDatasource{
		provider: p,
	}, diags
}

func (d groupDatasource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data model.GroupResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.Null {
		data.ID = data.Email
	}

	group := GetGroupData(&d.provider, &data, &resp.Diagnostics)
	if group.ID.Value == "" {
		resp.Diagnostics.AddError("object does not exist",
			fmt.Sprintf("Group %s does not exist", data.ID.Value))
	}

	diags = resp.State.Set(ctx, group)
	resp.Diagnostics.Append(diags...)
}

func GetGroupData(prov *provider, plan *model.GroupResourceData, diags *diag.Diagnostics) *model.GroupResourceData {
	groupsService := GetGroupsService(prov, diags)
	log.Printf("[DEBUG] Getting Group %s: %s", plan.ID.Value, plan.Email.Value)

	groupObj, err := groupsService.Get(plan.ID.Value).Do()
	if err != nil {
		plan.ID.Value = handleNotFoundError(err, plan.ID.Value, diags)
	}

	if groupObj == nil {
		diags.AddError("returned obj is nil", fmt.Sprintf("GET %s returned nil object",
			plan.ID.Value))
	}

	return SetGroupData(plan, groupObj)
}

func SetGroupData(plan *model.GroupResourceData, obj *directory.Group) *model.GroupResourceData {

	return &model.GroupResourceData{
		ID: types.String{
			Value: obj.Id,
		},
		Email: types.String{
			Value: obj.Email,
		},
		Name: types.String{
			Value: obj.Name,
		},
		Description: types.String{
			Value: obj.Description,
			Null:  plan.Description.Null,
		},
		AdminCreated: types.Bool{
			Value: obj.AdminCreated,
		},
		DirectMembersCount: types.Int64{
			Value: obj.DirectMembersCount,
		},
		Aliases:            stringSliceToTypeList(obj.Aliases),
		NonEditableAliases: stringSliceToTypeList(obj.Aliases),
	}
}

//import (
//	"context"
//	"fmt"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
//)
//
//func dataSourceGroup() *schema.Resource {
//	// Generate datasource schema from resource
//	dsSchema := datasourceSchemaFromResourceSchema(resourceGroup().Schema)
//	addExactlyOneOfFieldsToSchema(dsSchema, "id", "email")
//
//	return &schema.Resource{
//		// This description is used by the documentation generator and the language server.

//
//		ReadContext: dataSourceGroupRead,
//
//		Schema: dsSchema,
//	}
//}
//
//func dataSourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	if d.Get("id") != "" {
//		d.SetId(d.Get("id").(string))
//	} else {
//		var diags diag.Diagnostics
//
//		// use the meta value to retrieve your client from the provider configure method
//		client := meta.(*apiClient)
//
//		directoryService, diags := client.NewDirectoryService()
//		if diags.HasError() {
//			return diags
//		}
//
//		groupsService, diags := GetGroupsService(directoryService)
//		if diags.HasError() {
//			return diags
//		}
//
//		group, err := groupsService.Get(d.Get("email").(string)).Do()
//		if err != nil {
//			return diag.FromErr(err)
//		}
//
//		if group == nil {
//			diags = append(diags, diag.Diagnostic{
//				Severity: diag.Error,
//				Summary:  fmt.Sprintf("No group was returned for %s.", d.Get("email").(string)),
//			})
//
//			return diags
//		}
//
//		d.SetId(group.Id)
//	}
//
//	return resourceGroupRead(ctx, d, meta)
//}
