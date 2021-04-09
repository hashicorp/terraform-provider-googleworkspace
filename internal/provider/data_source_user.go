package googleworkspace

import (
	"context"
	"fmt"
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

		directoryService, diags := client.NewDirectoryService()
		if diags.HasError() {
			return diags
		}

		usersService, diags := GetUsersService(directoryService)
		if diags.HasError() {
			return diags
		}

		user, err := usersService.Get(d.Get("primary_email").(string)).Do()
		if err != nil {
			return diag.FromErr(err)
		}

		if user == nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("No user was returned for %s.", d.Get("primary_email").(string)),
			})

			return diags
		}

		d.SetId(user.Id)
	}

	return resourceUserRead(ctx, d, meta)
}
