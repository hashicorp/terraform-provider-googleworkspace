package googleworkspace

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/api/chromepolicy/v1"
)

func dataSourceChromePolicySchemas() *schema.Resource {

	chromePolicySchemaSchema := dataSourceChromePolicySchema().Schema
	chromePolicySchemaSchema["schema_name"].Computed = true
	chromePolicySchemaSchema["schema_name"].Required = false

	return &schema.Resource{
		Description: "Chrome Policy Schemas data source in the Terraform Googleworkspace provider.",

		ReadContext: dataSourceChromePolicySchemasRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Description: "The schema filter used to find a particular schema based on fields like its resource name, description and additionalTargetKeyNames.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"policy_schemas": {
				Description: "The list of policy schemas that match the query.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: chromePolicySchemaSchema,
				},
			},
		},
	}
}

func dataSourceChromePolicySchemasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	chromePolicyService, diags := client.NewChromePolicyService()
	if diags.HasError() {
		return diags
	}

	chromePolicySchemasService, diags := GetChromePolicySchemasService(chromePolicyService)

	req := chromePolicySchemasService.List("customers/" + client.Customer)

	filter := d.Get("filter").(string)
	if filter != "" {
		req = req.Filter(filter)
	}

	var policySchemas []interface{}
	err := req.Pages(ctx, func(resp *chromepolicy.GoogleChromePolicyV1ListPolicySchemasResponse) error {
		policySchemas = append(policySchemas, flattenPolicySchemas(resp.PolicySchemas)...)
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("policy_schemas", policySchemas); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func flattenPolicySchemas(ps []*chromepolicy.GoogleChromePolicyV1PolicySchema) []interface{} {
	result := make([]interface{}, len(ps))
	for i, p := range ps {
		obj := make(map[string]interface{})

		obj["schema_name"] = p.SchemaName
		obj["policy_description"] = p.PolicyDescription
		obj["support_uri"] = p.SupportUri
		obj["additional_target_key_names"] = flattenAdditionalTargetKeyNames(p.AdditionalTargetKeyNames)
		obj["definition"] = flattenDefinition(p.Definition)
		obj["access_restrictions"] = p.AccessRestrictions
		obj["notices"] = flattenNotices(p.Notices)

		result[i] = obj
	}
	return result
}
