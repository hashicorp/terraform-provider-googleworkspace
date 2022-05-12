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

type datasourceSchemaType struct{}

func (t datasourceSchemaType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	resourceType := resourceSchemaType{}

	resourceSchema, diags := resourceType.GetSchema(ctx)
	if diags.HasError() {
		return tfsdk.Schema{}, diags
	}

	attrs := datasourceSchemaFromResourceSchema(resourceSchema.Attributes)
	addExactlyOneOfFieldsToSchema(attrs, "schema_id", "schema_name")

	return tfsdk.Schema{
		Description: "Schema data source in the Terraform Googleworkspace provider. Schema resides " +
			"under the `https://www.googleapis.com/auth/admin.directory.userschema` client scope.",
		Attributes: attrs,
	}, nil
}

type schemaDatasource struct {
	provider provider
}

func (t datasourceSchemaType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	p, diags := convertProviderType(in)

	return schemaDatasource{
		provider: p,
	}, diags
}

func (d schemaDatasource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data model.SchemaResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.SchemaName.Null {
		data.SchemaName = data.SchemaId
	}

	schema := GetSchemaData(&d.provider, &data, &resp.Diagnostics)
	if schema.ID.Value == "" {
		resp.Diagnostics.AddError("object does not exist",
			fmt.Sprintf("Schema %s does not exist", data.SchemaName.Value))
	}

	diags = resp.State.Set(ctx, schema)
	resp.Diagnostics.Append(diags...)
}

func GetSchemaData(prov *provider, plan *model.SchemaResourceData, diags *diag.Diagnostics) *model.SchemaResourceData {
	schemasService := GetSchemasService(prov, diags)
	log.Printf("[DEBUG] Getting Schema %s", plan.SchemaName.Value)

	schemaObj, err := schemasService.Get(prov.customer, plan.SchemaName.Value).Do()
	if err != nil {
		plan.ID.Value = handleNotFoundError(err, plan.ID.Value, diags)
	}

	if schemaObj == nil {
		diags.AddError("returned obj is nil", fmt.Sprintf("GET %s returned nil object",
			plan.SchemaName.Value))
	}

	return SetSchemaData(plan, schemaObj)
}

func SetSchemaData(plan *model.SchemaResourceData, obj *directory.Schema) *model.SchemaResourceData {
	var fields types.List
	for _, f := range obj.Fields {

		field := model.SchemaResourceField{
			FieldName:           types.String{Value: f.FieldName},
			FieldId:             types.String{Value: f.FieldId},
			FieldType:           types.String{Value: f.FieldType},
			MultiValued:         types.Bool{Value: f.MultiValued},
			Indexed:             types.Bool{Value: f.Indexed},
			DisplayName:         types.String{Value: f.DisplayName},
			ReadAccessType:      types.String{Value: f.ReadAccessType},
			NumericIndexingSpec: numericIndexingSpec,
		}
		fields.Elems = append(fields.Elems, field)
	}

	return &model.SchemaResourceData{
		ID:          types.String{Value: obj.SchemaId},
		SchemaId:    types.String{Value: obj.SchemaId},
		SchemaName:  types.String{Value: obj.SchemaName},
		DisplayName: types.String{Value: obj.DisplayName},
		Fields:      fields,
	}
}

//func dataSourceSchema() *schema.Resource {
//	// Generate datasource schema from resource
//	dsSchema := datasourceSchemaFromResourceSchema(resourceSchema().Schema)
//	addExactlyOneOfFieldsToSchema(dsSchema, "schema_id", "schema_name")
//
//	return &schema.Resource{
//		// This description is used by the documentation generator and the language server.

//
//		ReadContext: dataSourceSchemaRead,
//
//		Schema: dsSchema,
//	}
//}
//
//func dataSourceSchemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	if d.Get("schema_id") != "" {
//		d.SetId(d.Get("schema_id").(string))
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
//		schemasService, diags := GetSchemasService(directoryService)
//		if diags.HasError() {
//			return diags
//		}
//
//		schema, err := schemasService.Get(client.Customer, d.Get("schema_name").(string)).Do()
//		if err != nil {
//			return diag.FromErr(err)
//		}
//
//		if schema == nil {
//			diags = append(diags, diag.Diagnostic{
//				Severity: diag.Error,
//				Summary:  fmt.Sprintf("No schema was returned for %s.", d.Get("schema_name").(string)),
//			})
//
//			return diags
//		}
//
//		d.SetId(schema.SchemaId)
//	}
//
//	return resourceSchemaRead(ctx, d, meta)
//}
