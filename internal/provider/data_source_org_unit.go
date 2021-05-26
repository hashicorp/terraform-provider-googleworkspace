package googleworkspace

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOrgUnit() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceOrgUnit().Schema)
	addExactlyOneOfFieldsToSchema(dsSchema, "org_unit_id", "org_unit_path")

	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "OrgUnit data source in the Terraform Googleworkspace provider.",

		ReadContext: dataSourceOrgUnitRead,

		Schema: dsSchema,
	}
}

func dataSourceOrgUnitRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Get("org_unit_id") != "" {
		d.SetId(d.Get("org_unit_id").(string))
	} else {
		var diags diag.Diagnostics

		// use the meta value to retrieve your client from the provider configure method
		client := meta.(*apiClient)

		directoryService, diags := client.NewDirectoryService()
		if diags.HasError() {
			return diags
		}

		orgUnitsService, diags := GetOrgUnitsService(directoryService)
		if diags.HasError() {
			return diags
		}

		orgUnit, err := orgUnitsService.Get(client.Customer, d.Get("name").(string)).Do()
		if err != nil {
			return diag.FromErr(err)
		}

		if orgUnit == nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("No org unit was returned for %s.", d.Get("org_unit_path").(string)),
			})

			return diags
		}

		d.SetId(orgUnit.OrgUnitId)
	}

	return resourceOrgUnitRead(ctx, d, meta)
}
