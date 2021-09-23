package googleworkspace

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGroupSettings() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceGroupSettings().Schema)
	addRequiredFieldsToSchema(dsSchema, "email")

	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Group Settings data source in the Terraform Googleworkspace provider. Group Settings resides " +
			"under the `https://www.googleapis.com/auth/apps.groups.settings` client scope.",

		ReadContext: dataSourceGroupSettingsRead,

		Schema: dsSchema,
	}
}

func dataSourceGroupSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	d.SetId(d.Get("email").(string))

	return resourceGroupSettingsRead(ctx, d, meta)
}
