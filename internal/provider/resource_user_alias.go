package googleworkspace

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	admin "google.golang.org/api/admin/directory/v1"
)

func resourceUserAlias() *schema.Resource {
	return &schema.Resource{
		Description:   "User alias resource manages individual aliases for the given Google workspace account.",
		CreateContext: resourceUserAliasCreate,
		ReadContext:   resourceUserAliasRead,
		DeleteContext: resourceUserAliasDelete,
		Importer: &schema.ResourceImporter{
			State: resourceUserAliasImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"primary_email": {
				Type:        schema.TypeString,
				Description: "Primary Email (userKey) of the user the alias should be applied to.",
				Required:    true,
				ForceNew:    true,
			},
			"alias": {
				Type:        schema.TypeString,
				Description: "Email alias which should be applied to the user.",
				Required:    true,
				ForceNew:    true,
			},
			"etag": {
				Description: "ETag of the resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			// Adding a computed id simply to override the `optional` id that gets added in the SDK
			// that will then display improperly in the docs
			"id": {
				Description: "The ID of this resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceUserAliasCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := meta.(*apiClient)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	usersService, diags := GetUsersService(directoryService)
	if diags.HasError() {
		return diags
	}

	aliasesService, diags := GetUserAliasService(usersService)
	if diags.HasError() {
		return diags
	}

	primaryEmail := d.Get("primary_email").(string)
	setAlias := d.Get("alias").(string)

	alias := &admin.Alias{
		Alias: setAlias,
	}
	alias, err := aliasesService.Insert(primaryEmail, alias).Do()
	if err != nil {
		return diag.Errorf("[ERROR] failed to add alias for user (%s): %v", primaryEmail, err)
	}

	// Using a different backoff style due to the fact the top level ETag is not a consistent value for individual aliases. In theory it will change when multiple
	// instances of this resource are used with the same primary email. The workaround is to wait until the list aliases call returns the actual alias created by the
	// specific instance of the module. Since there is no real update lifecycle for this type of thing, the ETag really is just there for self validation nothing changed,
	// in practice it should not matter too much.
	bOff := backoff.NewExponentialBackOff()
	bOff.MaxElapsedTime = d.Timeout(schema.TimeoutCreate)
	bOff.InitialInterval = time.Second

	err = backoff.Retry(func() error {
		resp, err := aliasesService.List(primaryEmail).Do()
		if err != nil {
			return backoff.Permanent(fmt.Errorf("[ERROR] could not retrieve aliases for user (%s): %v", primaryEmail, err))
		}

		_, ok := doesAliasExist(resp, setAlias)
		if ok {
			return nil
		}
		return fmt.Errorf("[WARN] no matching alias (%s) found for user (%s).", setAlias, primaryEmail)

	}, bOff)

	d.SetId(fmt.Sprintf("%s/%s", primaryEmail, alias.Alias))
	d.Set("primary_email", primaryEmail)
	d.Set("alias", alias.Alias)
	return resourceUserAliasRead(ctx, d, meta)
}

func resourceUserAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := meta.(*apiClient)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	usersService, diags := GetUsersService(directoryService)
	if diags.HasError() {
		return diags
	}

	aliasesService, diags := GetUserAliasService(usersService)
	if diags.HasError() {
		return diags
	}

	primaryEmail := d.Get("primary_email").(string)
	expectedAlias := d.Get("alias").(string)

	resp, err := aliasesService.List(primaryEmail).Do()
	if err != nil {
		return diag.Errorf("[ERROR] could not retrieve aliases for user (%s): %v", primaryEmail, err)
	}

	alias, ok := doesAliasExist(resp, expectedAlias)
	if !ok {
		log.Println(fmt.Sprintf("[WARN] no matching alias (%s) found for user (%s).", expectedAlias, primaryEmail))
		d.SetId("")
		return nil
	}
	d.SetId(fmt.Sprintf("%s/%s", alias.PrimaryEmail, alias.Alias))
	d.Set("primary_email", alias.PrimaryEmail)
	d.Set("alias", alias.Alias)
	d.Set("etag", alias.Etag)
	return nil
}

func resourceUserAliasDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := meta.(*apiClient)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	usersService, diags := GetUsersService(directoryService)
	if diags.HasError() {
		return diags
	}

	aliasesService, diags := GetUserAliasService(usersService)
	if diags.HasError() {
		return diags
	}

	primaryEmail := d.Get("primary_email").(string)
	alias := d.Get("alias").(string)

	err := aliasesService.Delete(primaryEmail, alias).Do()
	if err != nil {
		return diag.Errorf("[ERROR] unable to remove alias (%s) from user (%s): %v", alias, primaryEmail, err)
	}

	d.SetId("")
	return nil
}

func resourceUserAliasImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	primaryEmail := strings.Split(d.Id(), "/")[0]
	expectedAlias := strings.Split(d.Id(), "/")[1]

	d.SetId(fmt.Sprintf("%s/%s", primaryEmail, expectedAlias))
	d.Set("primary_email", primaryEmail)
	d.Set("alias", expectedAlias)

	return []*schema.ResourceData{d}, nil
}

func doesAliasExist(aliasesResp *admin.Aliases, expectedAlias string) (*admin.Alias, bool) {
	for _, aliasInt := range aliasesResp.Aliases {
		alias, ok := aliasInt.(map[string]interface{})
		if ok {
			if expectedAlias == safeInterfaceToString(alias["alias"]) {
				return &admin.Alias{
					Alias:        safeInterfaceToString(alias["alias"]),
					Etag:         safeInterfaceToString(alias["etag"]),
					Id:           safeInterfaceToString(alias["id"]),
					Kind:         safeInterfaceToString(alias["kind"]),
					PrimaryEmail: safeInterfaceToString(alias["primaryEmail"]),
				}, true
			}
		}
		if !ok {
			log.Println(fmt.Sprintf("[ERROR] alias format in response did not match sdk struct, this indicates a probelm with provider or sdk handling: %v", reflect.TypeOf(aliasInt)))
			return &admin.Alias{}, false
		}
	}
	return &admin.Alias{}, false
}
