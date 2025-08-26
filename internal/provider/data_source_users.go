// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	directory "google.golang.org/api/admin/directory/v1"
)

func dataSourceUsers() *schema.Resource {
	// Generate datasource schema from resource
	dsUserSchema := datasourceSchemaFromResourceSchema(resourceUser().Schema)

	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Users data source in the Terraform Googleworkspace provider. Users resides " +
			"under the `https://www.googleapis.com/auth/admin.directory.user` client scope.",

		ReadContext: dataSourceUsersRead,

		Schema: map[string]*schema.Schema{
			"users": {
				Description: "A list of User resources.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: dsUserSchema,
				},
			},
		},
	}
}

func dataSourceUsersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	usersService, diags := GetUsersService(directoryService)
	if diags.HasError() {
		return diags
	}

	var result []*directory.User
	err := usersService.List().Customer(client.Customer).Projection("full").Pages(ctx, func(resp *directory.Users) error {
		for _, user := range resp.Users {
			result = append(result, user)
		}

		return nil
	})

	if err != nil {
		return handleNotFoundError(err, d, "users")
	}

	if err := d.Set("users", flattenUsers(result, client)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("users")

	return diags
}

func flattenUsers(users []*directory.User, client *apiClient) interface{} {
	var result []interface{}

	for _, user := range users {
		result = append(result, flattenUser(user, client))
	}

	return result
}

func flattenUser(user *directory.User, client *apiClient) interface{} {
	var diags diag.Diagnostics

	customSchemas := []map[string]interface{}{}
	if len(user.CustomSchemas) > 0 {
		customSchemas, diags = flattenCustomSchemas(user.CustomSchemas, client)
		if diags.HasError() {
			return diags
		}
	}

	result := map[string]interface{}{}
	result["primary_email"] = user.PrimaryEmail
	result["is_admin"] = user.IsAdmin
	result["is_delegated_admin"] = user.IsDelegatedAdmin
	result["agreed_to_terms"] = user.AgreedToTerms
	result["suspended"] = user.Suspended
	result["change_password_at_next_login"] = user.ChangePasswordAtNextLogin
	result["ip_allowlist"] = user.IpWhitelisted
	result["name"] = flattenName(user.Name)
	result["emails"] = flattenInterfaceObjects(user.Emails)
	result["external_ids"] = flattenInterfaceObjects(user.ExternalIds)
	result["relations"] = flattenInterfaceObjects(user.Relations)
	result["aliases"] = user.Aliases
	result["is_mailbox_setup"] = user.IsMailboxSetup
	result["customer_id"] = user.CustomerId
	result["addresses"] = flattenInterfaceObjects(user.Addresses)
	result["organizations"] = flattenInterfaceObjects(user.Organizations)
	result["last_login_time"] = user.LastLoginTime
	result["phones"] = flattenInterfaceObjects(user.Phones)
	result["suspension_reason"] = user.SuspensionReason
	result["thumbnail_photo_url"] = user.ThumbnailPhotoUrl
	result["languages"] = flattenInterfaceObjects(user.Languages)
	result["posix_accounts"] = flattenInterfaceObjects(user.PosixAccounts)
	result["creation_time"] = user.CreationTime
	result["non_editable_aliases"] = user.NonEditableAliases
	result["ssh_public_keys"] = flattenInterfaceObjects(user.SshPublicKeys)
	result["websites"] = flattenInterfaceObjects(user.Websites)
	result["locations"] = flattenInterfaceObjects(user.Locations)
	result["include_in_global_address_list"] = user.IncludeInGlobalAddressList
	result["keywords"] = flattenInterfaceObjects(user.Keywords)
	result["deletion_time"] = user.DeletionTime
	result["thumbnail_photo_etag"] = user.ThumbnailPhotoEtag
	result["ims"] = flattenInterfaceObjects(user.Ims)
	result["custom_schemas"] = customSchemas
	result["is_enrolled_in_2_step_verification"] = user.IsEnrolledIn2Sv
	result["is_enforced_in_2_step_verification"] = user.IsEnforcedIn2Sv
	result["archived"] = user.Archived
	result["org_unit_path"] = user.OrgUnitPath
	result["recovery_email"] = user.RecoveryEmail
	result["recovery_phone"] = user.RecoveryPhone
	result["id"] = user.Id

	return result
}
