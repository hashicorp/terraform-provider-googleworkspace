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

		CreateContext: resourceChromePolicyCreate,
		UpdateContext: resourceChromePolicyUpdate,
		ReadContext:   resourceChromePolicyRead,
		DeleteContext: resourceChromePolicyDelete,

		Schema: map[string]*schema.Schema{
			"org_unit_id": {
				Description:      "The target org unit on which this policy is applied.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: diffSuppressOrgUnitId,
			},
			"policies": {
				Description: "Policies to set for the org unit",
				Type:        schema.TypeSet,
				Required:    true,
				// TODO: will need diffsuppressfunc
				// DiffSuppressFunc: nil,
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

func resourceChromePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	chromePolicyService, diags := client.NewChromePolicyService()
	if diags.HasError() {
		return diags
	}

	chromePoliciesService, diags := GetChromePoliciesService(chromePolicyService)
	if diags.HasError() {
		return diags
	}

	orgUnitId := strings.TrimPrefix(d.Get("org_unit_id").(string), "id:")

	log.Printf("[DEBUG] Creating Chrome Policy for org:%s", orgUnitId)

	policyTargetKey := &chromepolicy.GoogleChromePolicyV1PolicyTargetKey{
		TargetResource: "orgunits/" + orgUnitId,
	}
	policies := d.Get("policies").(*schema.Set).List()

	var modifyRequests []*chromepolicy.GoogleChromePolicyV1ModifyOrgUnitPolicyRequest

	for _, p := range policies {
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
		if err := json.Unmarshal(schemaValues, &valueMap); err != nil {
			return diag.FromErr(err)
		}
		for key := range valueMap {
			keys = append(keys, key)
		}
		updateMask := strings.Join(keys, ",")

		modifyRequests = append(modifyRequests, &chromepolicy.GoogleChromePolicyV1ModifyOrgUnitPolicyRequest{
			PolicyTargetKey: policyTargetKey,
			PolicyValue: &chromepolicy.GoogleChromePolicyV1PolicyValue{
				PolicySchema: schemaName,
				Value:        schemaValues,
			},
			UpdateMask: updateMask,
		})
	}

	_, err := chromePoliciesService.Orgunits.BatchModify(client.Customer, &chromepolicy.GoogleChromePolicyV1BatchModifyOrgUnitPoliciesRequest{Requests: modifyRequests}).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished creating Chrome Policy for org:%s", orgUnitId)
	d.SetId(orgUnitId)

	return resourceChromePolicyRead(ctx, d, meta)
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

	log.Printf("[DEBUG] Updating Chrome Policy for org:%s", d.Id())

	policyTargetKey := &chromepolicy.GoogleChromePolicyV1PolicyTargetKey{
		TargetResource: "orgunits/" + d.Id(),
	}

	// Update is achieved by inheriting defaults for the previous policySchemas, and then applying the new set
	old, _ := d.GetChange("policies")

	var requests []*chromepolicy.GoogleChromePolicyV1InheritOrgUnitPolicyRequest
	for _, p := range old.(*schema.Set).List() {
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

	// run create
	diags = resourceChromePolicyCreate(ctx, d, meta)

	log.Printf("[DEBUG] Finished Updating Chrome Policy for org:%s", d.Id())

	return diags
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

	log.Printf("[DEBUG] Getting Chrome Policy for org:%s", d.Id())

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

		if len(resp.ResolvedPolicies) != 1 {
			return diag.Errorf("unexpected nubmer of resolved policies for schema: %s", schemaName)
		}

		value := []byte(resp.ResolvedPolicies[0].Value.Value)
		policy["schema_values"] = string(value)

		policies[i] = policy
	}

	if err := d.Set("policies", policies); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished getting Chrome Policy for org:%s", d.Id())
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

	log.Printf("[DEBUG] Deleting Chrome Policy for org:%s", d.Id())

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

	log.Printf("[DEBUG] Finished deleting Chrome Policy for org:%s", d.Id())
	return nil
}
