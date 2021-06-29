package googleworkspace

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"google.golang.org/api/chromepolicy/v1"
)

func resourceChromePolicy() *schema.Resource {
	return &schema.Resource{
		Description: "Chrome Policy Schemas data source in the Terraform Googleworkspace provider.",

		CreateContext: resourceChromePolicyUpdate,
		UpdateContext: resourceChromePolicyUpdate,
		ReadContext:   resourceChromePolicyRead,
		DeleteContext: resourceChromePolicyDelete,

		Schema: map[string]*schema.Schema{
			"org_unit_id": {
				Description: "The target org unit on which this policy is applied.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				StateFunc:   stripOrgUnitId,
			},
			"policies": {
				Description: "Policies to set for the org unit",
				Type:        schema.TypeSet,
				Required:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"schema_name": {
							Description: "The full qualified name of the policy schema.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"schema_values": {
							Description:      "Values to be set for the chosen policy schema, must be encoded as JSON",
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsJSON),
						},
					},
				},
			},
		},
	}
}

func resourceChromePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	chromePolicyService, diags := client.NewChromePolicyService()
	if diags.HasError() {
		return diags
	}

	chromePoliciesService, diags := GetChromePoliciesService(chromePolicyService)
	if diags.HasError() {
		return diags
	}

	orgUnitId := d.Get("org_unit_id").(string)

	log.Printf("[DEBUG] Creating/Updating Chrome Policies for org:%s", orgUnitId)

	policyTargetKey := &chromepolicy.GoogleChromePolicyV1PolicyTargetKey{
		TargetResource: "orgunits/" + orgUnitId,
	}
	var requests []*chromepolicy.GoogleChromePolicyV1ModifyOrgUnitPolicyRequest

	for _, p := range d.Get("policies").(*schema.Set).List() {
		policy := p.(map[string]interface{})
		schemaName := policy["schema_name"].(string)
		schemaValues := []byte(policy["schema_values"].(string))

		// the object could in theory nest infinitely, but I'm assuming Google only
		// considers the "top level" as the fields to be set, especially since the
		// fields are qualified with dot notation. ex: chrome.printers.AllowUsers
		// One would think the structure should be flat, nested values are expressed
		// through the dot notation
		var valueMap map[string]interface{}
		var keys []string
		json.Unmarshal(schemaValues, &valueMap)
		for key := range valueMap {
			keys = append(keys, key)
		}

		requests = append(requests, &chromepolicy.GoogleChromePolicyV1ModifyOrgUnitPolicyRequest{
			PolicyTargetKey: policyTargetKey,
			PolicyValue: &chromepolicy.GoogleChromePolicyV1PolicyValue{
				PolicySchema: schemaName,
				Value:        schemaValues,
			},
			UpdateMask: strings.Join(keys, ","),
		})
	}

	_, err := chromePoliciesService.Orgunits.BatchModify(client.Customer, &chromepolicy.GoogleChromePolicyV1BatchModifyOrgUnitPoliciesRequest{Requests: requests}).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished creating/updating Chrome Policies for org:%s", orgUnitId)
	d.SetId(orgUnitId)

	return resourceChromePolicyRead(ctx, d, meta)
}

func resourceChromePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	chromePolicyService, diags := client.NewChromePolicyService()
	if diags.HasError() {
		return diags
	}

	chromePoliciesService, diags := GetChromePoliciesService(chromePolicyService)
	if diags.HasError() {
		return diags
	}

	log.Printf("[DEBUG] Getting Chrome Policies for org:%s", d.Id())

	policyTargetKey := &chromepolicy.GoogleChromePolicyV1PolicyTargetKey{
		TargetResource: "orgunits/" + d.Id(),
	}

	policies := d.Get("policies").(*schema.Set).List()
	for i, p := range policies {
		policy := p.(map[string]interface{})
		schemaName := policy["schema_name"].(string)
		// we will resolve each individual policySchema by fully qualified name, so the responses should be a single result
		resp, err := chromePoliciesService.Resolve(client.Customer, &chromepolicy.GoogleChromePolicyV1ResolveRequest{
			PolicySchemaFilter: schemaName,
			PolicyTargetKey:    policyTargetKey,
		}).Do()
		if err != nil {
			return diag.FromErr(err)
		}

		if len(resp.ResolvedPolicies) > 1 {
			return diag.Errorf("unexpected nubmer of resolved policies for schema: %s", schemaName)
		}

		value := []byte(resp.ResolvedPolicies[0].Value.Value)
		policy["schema_values"] = string(value)

		policies[i] = policy
	}

	if err := d.Set("policies", policies); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished getting Chrome Policies for org:%s", d.Id())
	return nil
}

func resourceChromePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	chromePolicyService, diags := client.NewChromePolicyService()
	if diags.HasError() {
		return diags
	}

	chromePoliciesService, diags := GetChromePoliciesService(chromePolicyService)
	if diags.HasError() {
		return diags
	}

	log.Printf("[DEBUG] Deleting Chrome Policies for org:%s", d.Id())

	policyTargetKey := &chromepolicy.GoogleChromePolicyV1PolicyTargetKey{
		TargetResource: "orgunits/" + d.Id(),
	}

	var requests []*chromepolicy.GoogleChromePolicyV1InheritOrgUnitPolicyRequest
	for _, p := range d.Get("policies").(*schema.Set).List() {
		policy := p.(map[string]interface{})
		schemaName := policy["schema_name"].(string)

		requests = append(requests, &chromepolicy.GoogleChromePolicyV1InheritOrgUnitPolicyRequest{
			PolicyTargetKey: policyTargetKey,
			PolicySchema:    schemaName,
		})
	}

	_, err := chromePoliciesService.Orgunits.BatchInherit(client.Customer, &chromepolicy.GoogleChromePolicyV1BatchInheritOrgUnitPoliciesRequest{Requests: requests}).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished deleting Chrome Policies for org:%s", d.Id())
	return nil
}
