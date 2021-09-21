package googleworkspace

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGroupMembers() *schema.Resource {
	// Generate datasource schema from resource
	dsSchema := datasourceSchemaFromResourceSchema(resourceGroupMembers().Schema)
	addRequiredFieldsToSchema(dsSchema, "group_id")

	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Group Members data source in the Terraform Googleworkspace provider.",

		ReadContext: dataSourceGroupMembersRead,

		Schema: dsSchema,
	}
}

func dataSourceGroupMembersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceGroupMembersRead(ctx, d, meta)
}
