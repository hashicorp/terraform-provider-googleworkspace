package googleworkspace

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/api/chromepolicy/v1"
)

func dataSourceChromePolicySchema() *schema.Resource {
	return &schema.Resource{
		Description: "Chrome Policy Schema data source in the Terraform Googleworkspace provider.",

		ReadContext: dataSourceChromePolicySchemaRead,

		Schema: map[string]*schema.Schema{
			// Intentionally ignoring field 'name' https://developers.google.com/chrome/policy/reference/rest/v1/customers.policySchemas#PolicySchema
			// it is a confusing field, that includes url segments the practitioner won't find useful.
			// Format: name=customers/{customer}/policySchemas/{schema_namespace}
			// Using the output field 'schema_name' instead as the field the practitioner specifies
			"schema_name": {
				Description: "The full qualified name of the policy schema",
				Type:        schema.TypeString,
				Required:    true,
			},
			"policy_description": {
				Description: "Description about the policy schema for user consumption.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"additional_target_key_names": {
				Description: "Additional key names that will be used to identify the target of the policy value. When specifying a policyTargetKey, each of the additional keys specified here will have to be included in the additionalTargetKeys map.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Description: "Key name.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"key_description": {
							Description: "Key description.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"definition": {
				Description: "Schema definition using proto descriptor.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "file name, relative to root of source tree",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"package": {
							Description: "e.g. 'foo', 'foo.bar', etc.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						// TODO: this field nests recursively
						// "message_type": {
						// 	Description: "All top-level definitions in this file.",
						// 	Type:        schema.TypeList,
						// 	Computed:    true,
						// 	Elem: &schema.Resource{
						// 		Schema: map[string]*schema.Schema{},
						// 	},
						// },
						"enum_type": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     enumDescriptorProto(),
						},
						"syntax": {
							Description: "The syntax of the proto file. The supported values are 'proto' and 'proto3'.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			// TODO: this field nests recursively
			// "field_descriptions": {
			// 	Description: "Detailed description of each field that is part of the schema.",
			// 	Type:        schema.TypeList,
			// 	Computed:    true,
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{},
			// 	},
			// },
			"access_restrictions": {
				Description: "Specific access restrictions related to this policy.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeList,
				},
			},
			"notices": {
				Description: "Special notice messages related to setting certain values in certain fields in the schema.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Description: "The field name associated with the notice.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"notice_value": {
							Description: "The value of the field that has a notice. When setting the field to this value, the user may be required to acknowledge the notice message in order for the value to be set.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"notice_message": {
							Description: "The notice message associate with the value of the field.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"acknowledgement_required": {
							Description: "Whether the user needs to acknowledge the notice message before the value can be set.",
							Type:        schema.TypeBool,
							Computed:    true,
						},
					},
				},
			},
			"support_uri": {
				Description: "URI to related support article for this schema.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func enumDescriptorProto() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"value": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"number": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceChromePolicySchemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	chromePolicyService, diags := client.NewChromePolicyService()
	if diags.HasError() {
		return diags
	}

	chromePolicySchemasService, diags := GetChromePolicySchemasService(chromePolicyService)
	if diags.HasError() {
		return diags
	}

	policySchema, err := chromePolicySchemasService.Get(fmt.Sprintf("customers/%s/policySchemas/%s", client.Customer, d.Get("schema_name").(string))).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(policySchema.SchemaName)
	d.Set("schema_name", policySchema.SchemaName)
	d.Set("policy_description", policySchema.PolicyDescription)
	d.Set("support_uri", policySchema.SupportUri)
	if err := d.Set("additional_target_key_names", flattenAdditionalTargetKeyNames(policySchema.AdditionalTargetKeyNames)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("definition", flattenDefinition(policySchema.Definition)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("access_restrictions", policySchema.AccessRestrictions); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("notices", flattenNotices(policySchema.Notices)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func flattenNotices(ns []*chromepolicy.GoogleChromePolicyV1PolicySchemaNoticeDescription) []interface{} {
	result := make([]interface{}, len(ns))

	for i, n := range ns {
		obj := make(map[string]interface{})
		obj["field"] = n.Field
		obj["notice_value"] = n.NoticeValue
		obj["notice_message"] = n.NoticeMessage
		obj["acknowledgement_required"] = n.AcknowledgementRequired
		result[i] = obj
	}

	return result
}

func flattenEnumType(es []*chromepolicy.Proto2EnumDescriptorProto) []interface{} {
	result := make([]interface{}, len(es))

	for i, e := range es {
		obj := make(map[string]interface{})

		obj["name"] = e.Name
		values := make([]interface{}, len(e.Value))
		for j, v := range e.Value {
			values[j] = map[string]interface{}{
				"name":   v.Name,
				"number": int(v.Number),
			}
		}
		obj["value"] = values

		result[i] = obj
	}

	return result
}

func flattenDefinition(d *chromepolicy.Proto2FileDescriptorProto) []interface{} {
	result := make([]interface{}, 1)
	obj := make(map[string]interface{})

	obj["name"] = d.Name
	obj["package"] = d.Package
	obj["syntax"] = d.Syntax
	obj["enum_type"] = flattenEnumType(d.EnumType)

	result[0] = obj
	return result
}

func flattenAdditionalTargetKeyNames(as []*chromepolicy.GoogleChromePolicyV1AdditionalTargetKeyName) []interface{} {
	result := make([]interface{}, len(as))
	for i, a := range as {
		obj := make(map[string]interface{})
		obj["key"] = a.Key
		obj["key_description"] = a.KeyDescription
		result[i] = obj
	}
	return result
}
