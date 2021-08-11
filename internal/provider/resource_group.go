package googleworkspace

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	directory "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

func resourceGroup() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Group resource manages Google Workspace Groups.",

		CreateContext: resourceGroupCreate,
		ReadContext:   resourceGroupRead,
		UpdateContext: resourceGroupUpdate,
		DeleteContext: resourceGroupDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique ID of a group. A group id can be used as a group request URI's groupKey.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"email": {
				Description: "The group's email address. If your account has multiple domains," +
					"select the appropriate domain for the email address. The email must be unique.",
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Description: "The group's display name.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Description: "An extended description to help users determine the purpose of a group." +
					"For example, you can include information about who should join the group," +
					"the types of messages to send to the group, links to FAQs about the group, or related groups.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(0, 4096)),
			},
			"admin_created": {
				Description: "Value is true if this group was created by an administrator rather than a user.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"direct_members_count": {
				Description: "The number of users that are direct members of the group." +
					"If a group is a member (child) of this group (the parent)," +
					"members of the child group are not counted in the directMembersCount property of the parent group.",
				Type:     schema.TypeInt,
				Computed: true,
			},
			"etag": {
				Description: "ETag of the resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"aliases": {
				Description: "asps.list of group's email addresses.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"non_editable_aliases": {
				Description: "asps.list of the group's non-editable alias email addresses that are outside of the" +
					"account's primary domain or subdomains. These are functioning email addresses used by the group.",
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	email := d.Get("email").(string)
	log.Printf("[DEBUG] Creating Group %q: %#v", email, email)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	groupsService, diags := GetGroupsService(directoryService)
	if diags.HasError() {
		return diags
	}

	groupObj := directory.Group{
		Email:       d.Get("email").(string),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	group, err := groupsService.Insert(&groupObj).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	// The etag changes with each insert, so we want to monitor how many changes we should see
	// when we're checking for eventual consistency
	numInserts := 1

	d.SetId(group.Id)

	aliases := d.Get("aliases.#").(int)

	if aliases > 0 {
		aliasesService, diags := GetGroupAliasService(groupsService)
		if diags.HasError() {
			return diags
		}

		for i := 0; i < aliases; i++ {
			aliasObj := directory.Alias{
				Alias: d.Get(fmt.Sprintf("aliases.%d", i)).(string),
			}

			_, err := aliasesService.Insert(d.Id(), &aliasObj).Do()
			if err != nil {
				return diag.FromErr(err)
			}
			numInserts += 1
		}
	}

	// INSERT will respond with the Group that will be created, however, it is eventually consistent
	// After INSERT, the etag is updated along with the Group (and any aliases),
	// once we get a consistent etag, we can feel confident that our Group is also consistent
	cc := consistencyCheck{
		resourceType: "group",
		timeout:      d.Timeout(schema.TimeoutCreate),
	}
	err = retryTimeDuration(ctx, d.Timeout(schema.TimeoutCreate), func() error {
		var retryErr error

		if cc.reachedConsistency(numInserts) {
			return nil
		}

		newGroup, retryErr := groupsService.Get(d.Id()).IfNoneMatch(cc.lastEtag).Do()
		if googleapi.IsNotModified(retryErr) {
			cc.currConsistent += 1
		} else if retryErr != nil {
			return fmt.Errorf("unexpected error during retries of %s: %s", cc.resourceType, retryErr)
		} else {
			cc.handleNewEtag(newGroup.Etag)
		}

		return fmt.Errorf("timed out while waiting for %s to be inserted", cc.resourceType)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished creating Group %q: %#v", d.Id(), email)

	return resourceGroupRead(ctx, d, meta)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	groupsService, diags := GetGroupsService(directoryService)
	if diags.HasError() {
		return diags
	}

	group, err := groupsService.Get(d.Id()).Do()
	if err != nil {
		return handleNotFoundError(err, d, d.Get("email").(string))
	}

	d.Set("email", group.Email)
	d.Set("name", group.Name)
	d.Set("description", group.Description)
	d.Set("admin_created", group.AdminCreated)
	d.Set("direct_members_count", group.DirectMembersCount)
	d.Set("aliases", group.Aliases)
	d.Set("non_editable_aliases", group.NonEditableAliases)
	d.Set("etag", group.Etag)

	d.SetId(group.Id)

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	email := d.Get("email").(string)
	log.Printf("[DEBUG] Updating Group %q: %#v", d.Id(), email)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	groupsService, diags := GetGroupsService(directoryService)
	if diags.HasError() {
		return diags
	}

	groupObj := directory.Group{}

	if d.HasChange("email") {
		groupObj.Email = d.Get("email").(string)
	}

	if d.HasChange("name") {
		groupObj.Name = d.Get("name").(string)
	}

	if d.HasChange("description") {
		groupObj.Description = d.Get("description").(string)
	}

	numInserts := 0
	if d.HasChange("aliases") {
		old, new := d.GetChange("aliases")
		oldAliases := listOfInterfacestoStrings(old.([]interface{}))
		newAliases := listOfInterfacestoStrings(new.([]interface{}))

		aliasesService, diags := GetGroupAliasService(groupsService)
		if diags.HasError() {
			return diags
		}

		// Remove old aliases that aren't in the new aliases list
		for _, alias := range oldAliases {
			if stringInSlice(newAliases, alias) {
				continue
			}

			err := aliasesService.Delete(d.Id(), alias).Do()
			if err != nil {
				return diag.FromErr(err)
			}
		}

		// Insert all new aliases that weren't previously in state
		for _, alias := range newAliases {
			if stringInSlice(oldAliases, alias) {
				continue
			}

			aliasObj := directory.Alias{
				Alias: alias,
			}

			_, err := aliasesService.Insert(d.Id(), &aliasObj).Do()
			if err != nil {
				return diag.FromErr(err)
			}
			numInserts += 1
		}
	}

	if &groupObj != new(directory.Group) {
		group, err := groupsService.Update(d.Id(), &groupObj).Do()
		if err != nil {
			return diag.FromErr(err)
		}
		numInserts += 1

		d.SetId(group.Id)
	}

	// UPDATE will respond with the Group that will be created, however, it is eventually consistent
	// After UPDATE, the etag is updated along with the Group (and any aliases),
	// once we get a consistent etag, we can feel confident that our Group is also consistent
	cc := consistencyCheck{
		resourceType: "group",
		timeout:      d.Timeout(schema.TimeoutUpdate),
	}
	err := retryTimeDuration(ctx, d.Timeout(schema.TimeoutUpdate), func() error {
		var retryErr error

		if cc.reachedConsistency(numInserts) {
			return nil
		}

		newGroup, retryErr := groupsService.Get(d.Id()).IfNoneMatch(cc.lastEtag).Do()
		if googleapi.IsNotModified(retryErr) {
			cc.currConsistent += 1
		} else if retryErr != nil {
			return fmt.Errorf("unexpected error during retries of %s: %s", cc.resourceType, retryErr)
		} else {
			cc.handleNewEtag(newGroup.Etag)
		}

		return fmt.Errorf("timed out while waiting for %s to be updated", cc.resourceType)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished creating Group %q: %#v", d.Id(), email)

	return resourceGroupRead(ctx, d, meta)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	email := d.Get("email").(string)
	log.Printf("[DEBUG] Deleting Group %q: %#v", d.Id(), email)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	groupsService, diags := GetGroupsService(directoryService)
	if diags.HasError() {
		return diags
	}

	err := groupsService.Delete(d.Id()).Do()
	if err != nil {
		return handleNotFoundError(err, d, d.Get("email").(string))
	}

	log.Printf("[DEBUG] Finished deleting Group %q: %#v", d.Id(), email)

	return diags
}
