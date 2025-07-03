package googleworkspace

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	directory "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

func resourceUserCustomSchemaAttributes() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "User Custom Schema Attributes resource manages Google Workspace " +
			"User Custom Schema Attributes. User resides under the " +
			"`https://www.googleapis.com/auth/admin.directory.user` client scope.",

		CreateContext: resourceUserCustomSchemaAttributesUpdate,
		ReadContext:   resourceUserCustomSchemaAttributesRead,
		UpdateContext: resourceUserCustomSchemaAttributesUpdate,
		DeleteContext: resourceUserCustomSchemaAttributesDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"primary_email": {
				Description: "The user's primary email address. The primaryEmail must be unique and cannot be an alias " +
					"of another user.",
				Type:     schema.TypeString,
				Required: true,
			},
			"custom_schemas": {
				Description:      "Schema Attributes of the user.",
				Type:             schema.TypeList,
				Required:         true,
				DiffSuppressFunc: diffSuppressCustomSchemas,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"schema_name": {
							Description: "The name of the schema.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"schema_values": {
							Description: "JSON encoded map that represents key/value pairs that " +
								"correspond to the given schema. ",
							Type:     schema.TypeMap,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateDiagFunc: validation.ToDiagFunc(
									validation.StringIsJSON,
								),
							},
						},
					},
				},
			},
		},

		CustomizeDiff: customdiff.All(
			customdiff.ForceNewIfChange("primary_email", func(ctx context.Context, old, new, meta interface{}) bool {
				return new.(string) != old.(string)
			}),
		),
	}
}

func resourceUserCustomSchemaAttributesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	primaryEmail := d.Get("primary_email").(string)

	log.Printf("[DEBUG] Updating Attributes of User: %#v", primaryEmail)
	if diags := updateUserCustomSchemaAttributes(ctx, d, meta, primaryEmail, false); diags.HasError() {
		return diags
	}
	log.Printf("[DEBUG] Finished updating Attributes of User: %#v", primaryEmail)

	return resourceUserCustomSchemaAttributesRead(ctx, d, meta)
}

func resourceUserCustomSchemaAttributesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	primaryEmail := d.Get("primary_email").(string)

	log.Printf("[DEBUG] Deleting Attributes of User: %#v", primaryEmail)
	if diags := updateUserCustomSchemaAttributes(ctx, d, meta, primaryEmail, true); diags.HasError() {
		return diags
	}
	log.Printf("[DEBUG] Finished deleting Attributes of User: %#v", primaryEmail)

	return resourceUserCustomSchemaAttributesRead(ctx, d, meta)
}

func resourceUserCustomSchemaAttributesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	primaryEmail := d.Get("primary_email").(string)
	log.Printf("[DEBUG] Getting Custom Schema Attributes of User %q: %#v", d.Id(), primaryEmail)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	usersService, diags := GetUsersService(directoryService)
	if diags.HasError() {
		return diags
	}

	userID, customSchemas, diags := getUserCustomSchemaAttributes(client, usersService, primaryEmail)
	if diags.HasError() {
		return diags
	}

	d.Set("custom_schemas", customSchemas)
	d.SetId(userID)
	log.Printf("[DEBUG] Finished getting Custom Schema Attributes of User %q: %#v", d.Id(), primaryEmail)

	return diags
}

func updateUserCustomSchemaAttributes(ctx context.Context, d *schema.ResourceData, meta interface{}, primaryEmail string, delete bool) diag.Diagnostics {
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

	customSchemas := make(map[string]googleapi.RawMessage)
	if len(d.Get("custom_schemas").([]interface{})) > 0 {
		diags := validateCustomSchemas(d, client)
		if diags.HasError() {
			return diags
		}

		customSchemas, diags = expandCustomSchemaValues(d.Get("custom_schemas").([]interface{}))
		if diags.HasError() {
			return diags
		}
	}

	_, rawCurrentCustomSchemas, diags := getUserCustomSchemaAttributes(client, usersService, primaryEmail)
	if diags.HasError() {
		return diags
	}

	updatedCustomSchemas := make(map[string]googleapi.RawMessage)
	if len(rawCurrentCustomSchemas) > 0 {
		for _, s := range rawCurrentCustomSchemas {
			parsedSchema := make(map[string]string)
			for k, v := range s["schema_values"].(map[string]interface{}) {
				parsedSchema[k] = v.(string)
			}

			values := make([]string, 0)
			for k := range parsedSchema {
				values = append(values, fmt.Sprintf("\"%v\":null", k))
			}

			updatedCustomSchemas[s["schema_name"].(string)] = googleapi.RawMessage(fmt.Sprintf("{%v}", strings.Join(values, ",")))
		}
	}

	if !delete {
		for k, v := range customSchemas {
			updatedCustomSchemas[k] = v
		}
	}

	_, err := usersService.Update(primaryEmail, &directory.User{
		PrimaryEmail:  primaryEmail,
		CustomSchemas: updatedCustomSchemas,
	}).Do()
	if err != nil {
		return diag.FromErr(err)
	}
	numInserts := 0

	// UPDATE will respond with the updated User, however, it is eventually consistent
	// After UPDATE, the etag is updated along with the User (and any aliases),
	// once we get a consistent etag, we can feel confident that our User is also consistent
	cc := consistencyCheck{
		resourceType: "user_attributes",
		timeout:      d.Timeout(schema.TimeoutUpdate),
	}
	if err := retryTimeDuration(ctx, d.Timeout(schema.TimeoutUpdate), func() error {
		var retryErr error

		if cc.reachedConsistency(numInserts) {
			return nil
		}

		newUser, retryErr := usersService.Get(primaryEmail).IfNoneMatch(cc.lastEtag).Do()
		if googleapi.IsNotModified(retryErr) {
			cc.currConsistent += 1
		} else if retryErr != nil {
			return fmt.Errorf("unexpected error during retries of %s: %s", cc.resourceType, retryErr)
		} else {
			cc.handleNewEtag(newUser.Etag)
		}

		return fmt.Errorf("timed out while waiting for %s to be updated", cc.resourceType)
	}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getUserCustomSchemaAttributes(client *apiClient, usersService *directory.UsersService, primaryEmail string) (string, []map[string]interface{}, diag.Diagnostics) {
	customSchemas := make([]map[string]interface{}, 0)

	user, err := usersService.Get(primaryEmail).Projection("full").Do()
	if err != nil {
		return "", customSchemas, diag.FromErr(err)
	}

	if user == nil {
		return "", customSchemas, diag.FromErr(errors.New(fmt.Sprintf("No user was returned for %s.", primaryEmail)))
	}

	if len(user.CustomSchemas) > 0 {
		var diags diag.Diagnostics
		customSchemas, diags = flattenCustomSchemas(user.CustomSchemas, client)
		if diags.HasError() {
			return "", make([]map[string]interface{}, 0), diags
		}
	}

	return user.Id, customSchemas, nil
}
