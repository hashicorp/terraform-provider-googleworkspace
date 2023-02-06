// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSchema() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceSchema().Schema)
	addExactlyOneOfFieldsToSchema(dsSchema, "schema_id", "schema_name")

	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Schema data source in the Terraform Googleworkspace provider. Schema resides " +
			"under the `https://www.googleapis.com/auth/admin.directory.userschema` client scope.",

		ReadContext: dataSourceSchemaRead,

		Schema: dsSchema,
	}
}

func dataSourceSchemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Get("schema_id") != "" {
		d.SetId(d.Get("schema_id").(string))
	} else {
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

		schema, err := schemasService.Get(client.Customer, d.Get("schema_name").(string)).Do()
		if err != nil {
			return diag.FromErr(err)
		}

		if schema == nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("No schema was returned for %s.", d.Get("schema_name").(string)),
			})

			return diags
		}

		d.SetId(schema.SchemaId)
	}

	return resourceSchemaRead(ctx, d, meta)
}
