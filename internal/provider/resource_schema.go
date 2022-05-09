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
)

type resourceSchemaType struct{}

// GetSchema Schema Resource
func (r resourceSchemaType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Schema resource manages Google Workspace Schemas. Schema resides " +
			"under the `https://www.googleapis.com/auth/admin.directory.userschema` client scope.",
		Attributes: map[string]tfsdk.Attribute{
			"schema_id": {
				Description: "The unique identifier of the schema.",
				Type:        types.StringType,
				Computed:    true,
			},
			"schema_name": {
				Description: "The schema's name.",
				Type:        types.StringType,
				Required:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
			},
			"display_name": {
				Description: "Display name for the schema.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"fields": {
				Description: "A list of fields in the schema.",
				Required:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"field_name": {
						Description: "The name of the field.",
						Type:        types.StringType,
						Required:    true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							tfsdk.RequiresReplace(),
						},
					},
					"field_id": {
						Description: "The unique identifier of the field.",
						Type:        types.StringType,
						Computed:    true,
					},
					"field_type": {
						Description: "The type of the field. Acceptable values are: " +
							"BOOL, DATE, DOUBLE, EMAIL, INT64, PHONE, STRING",
						Type:     types.StringType,
						Required: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							tfsdk.RequiresReplace(),
						},
						Validators: []tfsdk.AttributeValidator{
							stringInSliceValidator{
								stringOptions: []string{"BOOL", "DATE", "DOUBLE", "EMAIL", "INT64", "PHONE", "STRING"},
							},
						},
					},
					"multi_valued": {
						Description: "A boolean specifying whether this is a multi-valued field or not.",
						Type:        types.BoolType,
						Optional:    true,
						Computed:    true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							DefaultModifier{
								ValType:    types.BoolType,
								DefaultVal: false,
							},
						},
					},
					"indexed": {
						Description: "A boolean specifying whether the field is indexed or not.",
						Type:        types.BoolType,
						Optional:    true,
						Computed:    true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							DefaultModifier{
								ValType:    types.BoolType,
								DefaultVal: true,
							},
						},
					},
					"display_name": {
						Description: "Display Name of the field.",
						Type:        types.StringType,
						Optional:    true,
						Computed:    true,
					},
					"read_access_type": {
						Description: "Specifies who can view values of this field. " +
							"See Retrieve users as a non-administrator for more information. " +
							"Acceptable values are: ADMINS_AND_SELF or ALL_DOMAIN_USERS. " +
							"Note: It may take up to 24 hours for changes to this field to be reflected.",
						Type:     types.StringType,
						Optional: true,
						Computed: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							DefaultModifier{
								ValType:    types.StringType,
								DefaultVal: "ALL_DOMAIN_USERS",
							},
						},
						Validators: []tfsdk.AttributeValidator{
							stringInSliceValidator{
								stringOptions: []string{"ADMINS_AND_SELF", "ALL_DOMAIN_USERS"},
							},
						},
					},
					"numeric_indexing_spec": {
						Description: "Indexing spec for a numeric field. By default, " +
							"only exact match queries will be supported for numeric fields. " +
							"Setting the numericIndexingSpec allows range queries to be supported.",
						Optional: true,
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"min_value": {
								Description: "Minimum value of this field. This is meant to be indicative " +
									"rather than enforced. Values outside this range will still be indexed, " +
									"but search may not be as performant.",
								Type:     types.Float64Type,
								Optional: true,
							},
							"max_value": {
								Description: "Maximum value of this field. This is meant to be indicative " +
									"rather than enforced. Values outside this range will still be indexed, " +
									"but search may not be as performant.",
								Type:     types.Float64Type,
								Optional: true,
							},
						}),
					},
				}, tfsdk.ListNestedAttributesOptions{}),
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "Schema identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

type schemaResource struct {
	provider provider
}

func (r resourceSchemaType) NewResource(_ context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)

	return schemaResource{
		provider: p,
	}, diags
}

// Create a new schema
func (r schemaResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from "+
				"another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan model.SchemaResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaReq := SchemaPlanToObj(ctx, &plan, &resp.Diagnostics)

	log.Printf("[DEBUG] Creating Schema %s", plan.SchemaName.Value)
	schemasService := GetSchemasService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaObj, err := schemasService.Insert(r.provider.customer, &schemaReq).Do()
	if err != nil {
		resp.Diagnostics.AddError("error while trying to create schema", err.Error())
		return
	}

	if schemaObj == nil {
		resp.Diagnostics.AddError(fmt.Sprintf("no schema was returned for %s", plan.SchemaName.Value), "object returned was nil")
		return
	}

	schemaId := schemaObj.SchemaId
	numInserts := 1

	// INSERT will respond with the Schema that will be created, however, it is eventually consistent
	// After INSERT, the etag is updated along with the Schema,
	// once we get a consistent etag, we can feel confident that our Schema is also consistent
	cc := consistencyCheck{
		resourceType: "schema",
		timeout:      CreateTimeout,
	}
	err = retryTimeDuration(ctx, CreateTimeout, func() error {
		if cc.reachedConsistency(numInserts) {
			return nil
		}

		newOU, retryErr := schemasService.Get(r.provider.customer, schemaId).IfNoneMatch(cc.lastEtag).Do()
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

	plan.ID.Value = schemaId
	schema := GetSchemaData(&r.provider, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, schema)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Finished creating Schema %s: %s", schema.ID.Value, schema.SchemaName.Value)
}

// Read schema information
func (r schemaResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state schemaResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schema := GetSchemaData(&r.provider, state, &resp.Diagnostics)
	if schema.ID.Null {
		resp.State.RemoveResource(ctx)
		log.Printf("[DEBUG] Removed Schema from state because it was not found %s", state.ID.Value)
		return
	}

	diags = resp.State.Set(ctx, schema)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	log.Printf("[DEBUG] Finished getting Schema %s: %s", state.ID.Value, schema.SchemaName.Value)
}

// Update schema resource
func (r schemaResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Retrieve values from plan
	var plan schemaResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state schemaResourceData
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Updating Schema %q: %#v", plan.ID.Value, plan.SchemaName.Value)
	schemasService := GetSchemasService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	schemaReq := SchemaPlanToObj(ctx, &plan, &resp.Diagnostics)

	schemaObj, err := schemasService.Update(r.provider.customer, state.ID.Value, &schemaReq).Do()
	if err != nil {
		resp.Diagnostics.AddError("error while trying to create schema", err.Error())
		return
	}

	if schemaObj == nil {
		resp.Diagnostics.AddError(fmt.Sprintf("no schema was returned for %s", plan.SchemaName.Value), "object returned was nil")
		return
	}

	schemaId := schemaObj.SchemaId
	numInserts := 1

	// UPDATE will respond with the Schema that will be created, however, it is eventually consistent
	// After UPDATE, the etag is updated along with the Schema,
	// once we get a consistent etag, we can feel confident that our Schema is also consistent
	cc := consistencyCheck{
		resourceType: "schema",
		timeout:      UpdateTimeout,
	}
	err = retryTimeDuration(ctx, UpdateTimeout, func() error {
		if cc.reachedConsistency(numInserts) {
			return nil
		}

		newOU, retryErr := schemasService.Get(r.provider.customer, schemaId).IfNoneMatch(cc.lastEtag).Do()
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

	plan.ID.Value = schemaId
	schema := GetSchemaData(&r.provider, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, schema)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Finished creating Schema %q: %#v", state.ID.Value, plan.SchemaName.Value)
}

// Delete schema
func (r schemaResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state schemaResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Deleting Schema %q: %#v", state.ID.Value, state.SchemaName.Value)
	schemasService := GetSchemasService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := schemasService.Delete(r.provider.customer, state.ID.Value).Do()
	if err != nil {
		state.ID = types.String{Value: handleNotFoundError(err, state.ID.Value, &resp.Diagnostics)}
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.State.RemoveResource(ctx)
	log.Printf("[DEBUG] Finished deleting Schema %s: %s", state.ID.Value, state.SchemaName.Value)
}

// ImportState schema
func (r schemaResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

// Expand functions

func expandFields(ctx context.Context, fieldData types.List, diags *diag.Diagnostics) []*directory.SchemaFieldSpec {
	if len(fieldData.Elems) == 0 {
		return nil
	}

	fieldObjs := []*directory.SchemaFieldSpec{}

	for _, f := range fieldData.Elems {
		field := schemaFieldData{}
		d := f.(types.Object).As(ctx, &field, types.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return nil
		}

		numIndexSpec := schemaNumericIndexingSpecData{}
		d = field.NumericIndexingSpec.As(ctx, &numIndexSpec, types.ObjectAsOptions{
			UnhandledNullAsEmpty: true,
		})
		diags.Append(d...)
		if diags.HasError() {
			return nil
		}

		fieldObjs = append(fieldObjs, &directory.SchemaFieldSpec{
			FieldName:      field.FieldName.Value,
			FieldType:      field.FieldType.Value,
			MultiValued:    field.MultiValued.Value,
			Indexed:        &field.Indexed.Value,
			DisplayName:    field.DisplayName.Value,
			ReadAccessType: field.ReadAccessType.Value,
			NumericIndexingSpec: &directory.SchemaFieldSpecNumericIndexingSpec{
				MinValue: numIndexSpec.MinValue.Value,
				MaxValue: numIndexSpec.MaxValue.Value,
			},
		})
	}

	return fieldObjs
}
