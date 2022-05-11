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

type datasourceOrgUnitType struct{}

func (t datasourceOrgUnitType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	resourceType := resourceOrgUnitType{}

	resourceSchema, diags := resourceType.GetSchema(ctx)
	if diags.HasError() {
		return tfsdk.Schema{}, diags
	}

	attrs := datasourceSchemaFromResourceSchema(resourceSchema.Attributes)
	addExactlyOneOfFieldsToSchema(attrs, "org_unit_id", "org_unit_path")

	return tfsdk.Schema{
		Description: "Org Unit data source in the Terraform Googleworkspace provider. Org Unit resides " +
			"under the `https://www.googleapis.com/auth/admin.directory.orgunit` client scope.",
		Attributes: attrs,
	}, nil
}

type orgUnitDatasource struct {
	provider provider
}

func (t datasourceOrgUnitType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	p, diags := convertProviderType(in)

	return orgUnitDatasource{
		provider: p,
	}, diags
}

func (d orgUnitDatasource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data model.OrgUnitResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.OrgUnitId.Null {
		data.OrgUnitId = data.OrgUnitPath
	}

	orgUnit := GetOrgUnitData(&d.provider, &data, &resp.Diagnostics)
	if orgUnit.ID.Value == "" {
		resp.Diagnostics.AddError("object does not exist",
			fmt.Sprintf("Org Unit %s does not exist", data.Name.Value))
	}

	diags = resp.State.Set(ctx, orgUnit)
	resp.Diagnostics.Append(diags...)
}

func GetOrgUnitData(prov *provider, plan *model.OrgUnitResourceData, diags *diag.Diagnostics) *model.OrgUnitResourceData {
	orgUnitsService := GetOrgUnitsService(prov, diags)
	log.Printf("[DEBUG] Getting Org Unit %s: %s", GetOrgUnitId(plan.OrgUnitId.Value), plan.Name.Value)

	orgUnitObj, err := orgUnitsService.Get(prov.customer, plan.OrgUnitId.Value).Do()
	if err != nil {
		plan.ID.Value = handleNotFoundError(err, plan.ID.Value, diags)
	}

	if orgUnitObj == nil {
		diags.AddError("returned obj is nil", fmt.Sprintf("GET %s returned nil object",
			plan.OrgUnitId.Value))
	}

	return SetOrgUnitData(plan, orgUnitObj)
}

func SetOrgUnitData(plan *model.OrgUnitResourceData, obj *directory.OrgUnit) *model.OrgUnitResourceData {
	return &model.OrgUnitResourceData{
		ID: types.String{
			Value: GetOrgUnitId(obj.OrgUnitId),
		},
		Name: types.String{
			Value: obj.Name,
		},
		Description: types.String{
			Value: obj.Description,
			Null:  plan.Description.Null,
		},
		BlockInheritance: types.Bool{
			Value: obj.BlockInheritance,
		},
		OrgUnitId: types.String{
			Value: obj.OrgUnitId,
		},
		OrgUnitPath: types.String{
			Value: obj.OrgUnitPath,
		},
		ParentOrgUnitId: types.String{
			Value: obj.ParentOrgUnitId,
		},
		ParentOrgUnitPath: types.String{
			Value: obj.ParentOrgUnitPath,
		},
	}
}

//func dataSourceOrgUnitRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	if d.Get("org_unit_id") != "" {
//		d.SetId(d.Get("org_unit_id").(string))
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
//		orgUnitsService, diags := GetOrgUnitsService(directoryService)
//		if diags.HasError() {
//			return diags
//		}
//
//		orgUnitPath := d.Get("org_unit_path").(string)
//		ouPath := strings.TrimLeft(orgUnitPath, "/")
//
//		orgUnit, err := orgUnitsService.Get(client.Customer, ouPath).Do()
//		if err != nil {
//			return diag.FromErr(err)
//		}
//
//		if orgUnit == nil {
//			diags = append(diags, diag.Diagnostic{
//				Severity: diag.Error,
//				Summary:  fmt.Sprintf("No org unit was returned for %s.", orgUnitPath),
//			})
//
//			return diags
//		}
//
//		d.SetId(orgUnit.OrgUnitId)
//	}
//
//	return resourceOrgUnitRead(ctx, d, meta)
//}
