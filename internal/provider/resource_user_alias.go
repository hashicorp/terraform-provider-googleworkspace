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
		Description:   "User Alias resources manages individual aliases for the given Google workspace account.",
		CreateContext: resourceUserAliasCreate,
		ReadContext:   resourceUserAliasRead,
		UpdateContext: nil,
		DeleteContext: resourceUserAliasDelete,
		Importer: &schema.ResourceImporter{
			State: resourceUserAliasImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
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
	_, err := aliasesService.Insert(primaryEmail, alias).Do()
	if err != nil {
		return diag.Errorf("[ERROR] failed to add alias for user (%s): %v", primaryEmail, err)
	}

	bOff := backoff.NewExponentialBackOff()
	bOff.MaxElapsedTime = d.Timeout(schema.TimeoutUpdate)
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

	d.SetId(fmt.Sprintf("%s/%s", alias.PrimaryEmail, alias.Alias))
	d.Set("primary_email", alias.PrimaryEmail)
	d.Set("alias", alias.Alias)
	d.Set("etag", alias.Etag)
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
	var diags diag.Diagnostics

	client := meta.(*apiClient)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return nil, fmt.Errorf("[ERROR] Unable to init client: %s", diags[0].Summary)
	}

	usersService, diags := GetUsersService(directoryService)
	if diags.HasError() {
		return nil, fmt.Errorf("[ERROR] Unable to init client: %s", diags[0].Summary)
	}

	aliasesService, diags := GetUserAliasService(usersService)
	if diags.HasError() {
		return nil, fmt.Errorf("[ERROR] Unable to init client: %s", diags[0].Summary)
	}

	primaryEmail := strings.Split(d.Id(), "/")[0]
	expectedAlias := strings.Split(d.Id(), "/")[1]

	resp, err := aliasesService.List(primaryEmail).Do()
	if err != nil {
		return nil, fmt.Errorf("[ERROR] could not retrieve aliases for user (%s): %v", primaryEmail, err)
	}

	alias, ok := doesAliasExist(resp, expectedAlias)
	if !ok {
		return nil, fmt.Errorf("[ERROR] no matching alias (%s) found for user (%s).", expectedAlias, primaryEmail)
	}
	d.SetId(fmt.Sprintf("%s/%s", alias.PrimaryEmail, alias.Alias))
	d.Set("primary_email", alias.PrimaryEmail)
	d.Set("alias", alias.Alias)

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
					PrimaryEmail: safeInterfaceToString(alias["primaryemail"]),
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
