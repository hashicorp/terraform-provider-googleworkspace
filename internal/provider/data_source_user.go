package googleworkspace

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUser() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceUser().Schema)
	addExactlyOneOfFieldsToSchema(dsSchema, "id", "primary_email")

	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "User data source in the Terraform Googleworkspace provider.",

		ReadContext: dataSourceUserRead,

		Schema: dsSchema,
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Get("id") != "" {
		d.SetId(d.Get("id").(string))
	} else {
		var diags diag.Diagnostics

		// use the meta value to retrieve your client from the provider configure method
		client := meta.(*apiClient)

		directoryService := client.NewDirectoryService()
		if directoryService == nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Directory Service could not be created.",
			})

			return diags
		}

		user, err := directoryService.Users.Get(d.Get("primary_email").(string)).Do()
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(user.Id)
	}

	return resourceUserRead(ctx, d, meta)
}
