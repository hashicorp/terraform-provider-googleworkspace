package googleworkspace

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	directory "google.golang.org/api/admin/directory/v1"
)

func resourceSchema() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Schema resource manages Google Workspace Schemas. Schema resides " +
			"under the `https://www.googleapis.com/auth/admin.directory.userschema` client scope.",

		CreateContext: resourceSchemaCreate,
		ReadContext:   resourceSchemaRead,
		UpdateContext: resourceSchemaUpdate,
		DeleteContext: resourceSchemaDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"schema_id": {
				Description: "The unique identifier of the schema.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"schema_name": {
				Description: "The schema's name.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"fields": {
				Description: "A list of fields in the schema.",
				Type:        schema.TypeList,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field_name": {
							Description: "The name of the field.",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
						"field_id": {
							Description: "The unique identifier of the field.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"field_type": {
							Description: "The type of the field. Acceptable values are: " +
								"BOOL, DATE, DOUBLE, EMAIL, INT64, PHONE, STRING",
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
								"BOOL", "DATE", "DOUBLE", "EMAIL", "INT64", "PHONE", "STRING"}, true)),
						},
						"multi_valued": {
							Description: "A boolean specifying whether this is a multi-valued field or not.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"etag": {
							Description: "The ETag of the field.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"indexed": {
							Description: "Boolean specifying whether the field is indexed or not.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
						},
						"display_name": {
							Description: "Display Name of the field.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
						},
						"read_access_type": {
							Description: "Specifies who can view values of this field. " +
								"See Retrieve users as a non-administrator for more information. " +
								"Acceptable values are: ADMINS_AND_SELF or ALL_DOMAIN_USERS. " +
								"Note: It may take up to 24 hours for changes to this field to be reflected.",
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "ALL_DOMAIN_USERS",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"ADMINS_AND_SELF", "ALL_DOMAIN_USERS"}, true)),
						},
						// TODO: (mbang) AtLeastOneOf (https://github.com/hashicorp/terraform-plugin-sdk/issues/470)
						"numeric_indexing_spec": {
							Description: "Indexing spec for a numeric field. By default, " +
								"only exact match queries will be supported for numeric fields. " +
								"Setting the numericIndexingSpec allows range queries to be supported.",
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"min_value": {
										Description: "Minimum value of this field. This is meant to be indicative " +
											"rather than enforced. Values outside this range will still be indexed, " +
											"but search may not be as performant.",
										Type:     schema.TypeFloat,
										Optional: true,
									},
									"max_value": {
										Description: "Maximum value of this field. This is meant to be indicative " +
											"rather than enforced. Values outside this range will still be indexed, " +
											"but search may not be as performant.",
										Type:     schema.TypeFloat,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"display_name": {
				Description: "Display name for the schema.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"etag": {
				Description: "ETag of the resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			// Adding a computed id simply to override the `optional` id that gets added in the SDK
			// that will then display improperly in the docs
			"id": {
				Description: "The ID of this resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceSchemaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	schemaName := d.Get("schema_name").(string)
	log.Printf("[DEBUG] Creating Schema %q: %#v", d.Id(), schemaName)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	schemasService, diags := GetSchemasService(directoryService)
	if diags.HasError() {
		return diags
	}

	schemaObj := directory.Schema{
		SchemaName:  d.Get("schema_name").(string),
		Fields:      expandFields(d.Get("fields")),
		DisplayName: d.Get("display_name").(string),
	}

	err := retryTimeDuration(ctx, d.Timeout(schema.TimeoutCreate), func() error {
		definedSchema, retryErr := schemasService.Insert(client.Customer, &schemaObj).Do()
		if retryErr != nil {
			return retryErr
		}

		d.SetId(definedSchema.SchemaId)
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished creating Schema %q: %#v", d.Id(), schemaName)

	return resourceSchemaRead(ctx, d, meta)
}

func resourceSchemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	schemasService, diags := GetSchemasService(directoryService)
	if diags.HasError() {
		return diags
	}

	schemaName := d.Get("schema_name").(string)
	log.Printf("[DEBUG] Getting Schema %q: %#v", d.Id(), schemaName)

	schema, err := schemasService.Get(client.Customer, d.Id()).Do()
	if err != nil {
		return handleNotFoundError(err, d, schemaName)
	}

	if schema == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("No schema was returned for %s.", d.Get("schema_name").(string)),
		})

		return diags
	}

	d.Set("schema_id", schema.SchemaId)
	d.Set("schema_name", schema.SchemaName)
	d.Set("fields", flattenFields(schema.Fields))
	d.Set("display_name", schema.DisplayName)
	d.Set("etag", schema.Etag)
	d.SetId(schema.SchemaId)
	log.Printf("[DEBUG] Finished getting Schema %q: %#v", d.Id(), schemaName)

	return diags
}

func resourceSchemaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	schemaName := d.Get("schema_name").(string)
	log.Printf("[DEBUG] Updating Schema %q: %#v", d.Id(), schemaName)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	schemasService, diags := GetSchemasService(directoryService)
	if diags.HasError() {
		return diags
	}

	schemaObj := directory.Schema{}

	// Strings

	if d.HasChange("schema_name") {
		schemaObj.SchemaName = schemaName
	}

	if d.HasChange("fields") {
		schemaObj.Fields = expandFields(d.Get("fields"))
	}

	if d.HasChange("display_name") {
		schemaObj.DisplayName = d.Get("display_name").(string)
	}

	if &schemaObj != new(directory.Schema) {
		schemaObj.SchemaId = d.Id()

		err := retryTimeDuration(ctx, d.Timeout(schema.TimeoutUpdate), func() error {
			definedSchema, retryErr := schemasService.Update(client.Customer, d.Id(), &schemaObj).Do()
			if retryErr != nil {
				return retryErr
			}

			d.SetId(definedSchema.SchemaId)
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	log.Printf("[DEBUG] Finished updating Schema %q: %#v", d.Id(), schemaName)

	return resourceSchemaRead(ctx, d, meta)
}

func resourceSchemaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	schemaName := d.Get("schema_name").(string)
	log.Printf("[DEBUG] Deleting Schema %q: %#v", d.Id(), schemaName)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	schemasService, diags := GetSchemasService(directoryService)
	if diags.HasError() {
		return diags
	}

	err := retryTimeDuration(ctx, d.Timeout(schema.TimeoutDelete), func() error {
		retryErr := schemasService.Delete(client.Customer, d.Id()).Do()
		if retryErr != nil {
			return retryErr
		}

		return nil
	})
	if err != nil {
		return handleNotFoundError(err, d, schemaName)
	}

	log.Printf("[DEBUG] Finished deleting Schema %q: %#v", d.Id(), schemaName)

	return diags
}

// Expand functions

func expandFields(v interface{}) []*directory.SchemaFieldSpec {
	fields := v.([]interface{})

	if len(fields) == 0 {
		return nil
	}

	fieldObjs := []*directory.SchemaFieldSpec{}

	for _, field := range fields {
		fieldObjs = append(fieldObjs, &directory.SchemaFieldSpec{
			FieldName:           field.(map[string]interface{})["field_name"].(string),
			FieldType:           field.(map[string]interface{})["field_type"].(string),
			MultiValued:         field.(map[string]interface{})["multi_valued"].(bool),
			Indexed:             expandNestedFieldsIndexed(field.(map[string]interface{})["indexed"]),
			DisplayName:         field.(map[string]interface{})["display_name"].(string),
			ReadAccessType:      field.(map[string]interface{})["read_access_type"].(string),
			NumericIndexingSpec: expandNestedNumericIndexingSpec(field.(map[string]interface{})["numeric_indexing_spec"]),
		})
	}

	return fieldObjs
}

func expandNestedNumericIndexingSpec(v interface{}) *directory.SchemaFieldSpecNumericIndexingSpec {
	numericIndexingSpec := v.([]interface{})
	numericIndexingSpecObj := directory.SchemaFieldSpecNumericIndexingSpec{}

	if len(numericIndexingSpec) > 0 {
		numericIndexingSpecObj.MinValue = numericIndexingSpec[0].(map[string]interface{})["min_value"].(float64)
		numericIndexingSpecObj.MaxValue = numericIndexingSpec[0].(map[string]interface{})["max_value"].(float64)
	}

	return &numericIndexingSpecObj

}

func expandNestedFieldsIndexed(v interface{}) *bool {
	if v == nil {
		return nil
	}

	indexed := v.(bool)

	return &indexed
}

// Flatten functions

func flattenFields(fieldObjs []*directory.SchemaFieldSpec) interface{} {
	fields := []map[string]interface{}{}

	for _, fieldObj := range fieldObjs {
		fields = append(fields, map[string]interface{}{
			"field_name":            fieldObj.FieldName,
			"field_id":              fieldObj.FieldId,
			"field_type":            fieldObj.FieldType,
			"multi_valued":          fieldObj.MultiValued,
			"etag":                  fieldObj.Etag,
			"indexed":               flattenNestedSchemaIndexed(fieldObj.Indexed),
			"display_name":          fieldObj.DisplayName,
			"read_access_type":      fieldObj.ReadAccessType,
			"numeric_indexing_spec": flattenNestedSchemaNumericIndexingSpec(fieldObj.NumericIndexingSpec),
		})
	}

	return fields
}

func flattenNestedSchemaIndexed(indexed *bool) interface{} {
	if indexed == nil {
		return nil
	}

	return indexed
}

func flattenNestedSchemaNumericIndexingSpec(numericIndexingSpecObj *directory.SchemaFieldSpecNumericIndexingSpec) interface{} {
	if numericIndexingSpecObj == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"min_value": numericIndexingSpecObj.MinValue,
			"max_value": numericIndexingSpecObj.MaxValue,
		},
	}
}
