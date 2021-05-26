package googleworkspace

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOrgUnit() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceOrgUnit().Schema)
	addRequiredFieldsToSchema(dsSchema, "org_unit_id")

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
	}

	return resourceOrgUnitRead(ctx, d, meta)
}
