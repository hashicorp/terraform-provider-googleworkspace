// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	directory "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

func resourceGroupMember() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Group Member resource manages Google Workspace Groups Members. Group Member resides under the " +
			"`https://www.googleapis.com/auth/admin.directory.group` client scope.",

		CreateContext: resourceGroupMemberCreate,
		ReadContext:   resourceGroupMemberRead,
		UpdateContext: resourceGroupMemberUpdate,
		DeleteContext: resourceGroupMemberDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: resourceGroupMemberImport,
		},

		Schema: map[string]*schema.Schema{
			"group_id": {
				Description: "Identifies the group in the API request. The value can be the group's email address, " +
					"group alias, or the unique group ID.",
				Type:     schema.TypeString,
				Required: true,
			},
			"email": {
				Description: "The member's email address. A member can be a user or another group. This property is " +
					"required when adding a member to a group. The email must be unique and cannot be an alias of " +
					"another group. If the email address is changed, the API automatically reflects the email address changes.",
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"role": {
				Description: "The member's role in a group. The API returns an error for cycles in group memberships. " +
					"For example, if group1 is a member of group2, group2 cannot be a member of group1. " +
					"Acceptable values are:" +
					"\n\t- `MANAGER`: This role is only available if the Google Groups for Business is " +
					"enabled using the Admin Console. A `MANAGER` role can do everything done by an `OWNER` role except " +
					"make a member an `OWNER` or delete the group. A group can have multiple `MANAGER` members. " +
					"\n\t- `MEMBER`: This role can subscribe to a group, view discussion archives, and view the group's " +
					"membership list." +
					"\n\t- `OWNER`: This role can send messages to the group, add or remove members, change member roles, " +
					"change group's settings, and delete the group. An OWNER must be a member of the group. " +
					"A group can have more than one OWNER.",
				Type:     schema.TypeString,
				Optional: true,
				Default:  "MEMBER",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"MANAGER", "MEMBER", "OWNER"},
					false)),
			},
			"etag": {
				Description: "ETag of the resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"type": {
				Description: "The type of group member. Acceptable values are: " +
					"\n\t- `CUSTOMER`: The member represents all users in a domain. An email address is not returned and the " +
					"ID returned is the customer ID. " +
					"\n\t- `GROUP`: The member is another group. " +
					"\n\t- `USER`: The member is a user.",
				Type:     schema.TypeString,
				Optional: true,
				Default:  "USER",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"CUSTOMER", "GROUP", "USER"},
					false)),
			},
			"status": {
				Description: "Status of member.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"delivery_settings": {
				Description: "Defines mail delivery preferences of member. Acceptable values are: " +
					"\n\t- `ALL_MAIL`: All messages, delivered as soon as they arrive." +
					"\n\t- `DAILY`: No more than one message a day." +
					"\n\t- `DIGEST`: Up to 25 messages bundled into a single message." +
					"\n\t- `DISABLED`: Remove subscription." +
					"\n\t- `NONE`: No messages.",
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ALL_MAIL",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"ALL_MAIL", "DAILY", "DIGEST",
					"DISABLED", "NONE"}, false)),
			},
			"member_id": {
				Description: "The unique ID of the group member. A member id can be used as a member request URI's memberKey.",
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

func resourceGroupMemberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	email := d.Get("email").(string)
	groupId := d.Get("group_id").(string)
	log.Printf("[DEBUG] Creating Group Member %q in groupu %s: %#v", email, groupId, email)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	membersService, diags := GetMembersService(directoryService)
	if diags.HasError() {
		return diags
	}

	memberObj := directory.Member{
		Email:            d.Get("email").(string),
		Role:             d.Get("role").(string),
		Type:             d.Get("type").(string),
		DeliverySettings: d.Get("delivery_settings").(string),
	}

	member, err := membersService.Insert(groupId, &memberObj).Do()

	// If we receive a 409 that the member already exists, ignore it, we'll import it next
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("member_id", member.Id)
	d.SetId(fmt.Sprintf("groups/%s/members/%s", groupId, member.Id))

	// INSERT will respond with the Group Member that will be created, however, it is eventually consistent
	// After INSERT, the etag is updated along with the Group Member,
	// once we get a consistent etag, we can feel confident that our Group Member is also consistent
	cc := consistencyCheck{
		resourceType: "group_member",
		timeout:      d.Timeout(schema.TimeoutCreate),
	}
	err = retryTimeDuration(ctx, d.Timeout(schema.TimeoutCreate), func() error {
		var retryErr error

		if cc.reachedConsistency(1) {
			return nil
		}

		newMember, retryErr := membersService.Get(groupId, member.Id).IfNoneMatch(cc.lastEtag).Do()
		if googleapi.IsNotModified(retryErr) {
			cc.currConsistent += 1
		} else if isNotFound(retryErr) {
			// groupmember was not found yet therefore setting currConsistent back to null value
			cc.currConsistent = 0
		} else if retryErr != nil {
			return fmt.Errorf("unexpected error during retries of %s: %s", cc.resourceType, retryErr)
		} else {
			cc.handleNewEtag(newMember.Etag)
		}

		return fmt.Errorf("timed out while waiting for %s to be inserted", cc.resourceType)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished creating Group Member %q: %#v", member.Id, email)

	return resourceGroupMemberRead(ctx, d, meta)
}

func resourceGroupMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	membersService, diags := GetMembersService(directoryService)
	if diags.HasError() {
		return diags
	}

	groupId := d.Get("group_id").(string)
	memberId := d.Get("member_id").(string)

	member, err := membersService.Get(groupId, memberId).Do()
	if err != nil {
		return handleNotFoundError(err, d, d.Id())
	}

	d.Set("email", member.Email)
	d.Set("role", member.Role)
	d.Set("etag", member.Etag)
	d.Set("type", member.Type)
	d.Set("status", member.Status)
	d.Set("delivery_settings", member.DeliverySettings)
	d.Set("member_id", member.Id)

	d.SetId(fmt.Sprintf("groups/%s/members/%s", groupId, member.Id))

	return diags
}

func resourceGroupMemberUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	email := d.Get("email").(string)
	memberId := d.Get("member_id").(string)
	log.Printf("[DEBUG] Updating Group Member %q: %#v", memberId, email)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	membersService, diags := GetMembersService(directoryService)
	if diags.HasError() {
		return diags
	}

	memberObj := directory.Member{}

	if d.HasChange("email") {
		memberObj.Email = d.Get("email").(string)
	}

	if d.HasChange("role") {
		memberObj.Role = d.Get("role").(string)
	}

	if d.HasChange("type") {
		memberObj.Type = d.Get("type").(string)
	}

	if d.HasChange("delivery_settings") {
		memberObj.DeliverySettings = d.Get("delivery_settings").(string)
	}

	if &memberObj != new(directory.Member) {
		groupId := d.Get("group_id").(string)
		memberId := d.Get("member_id").(string)
		member, err := membersService.Update(groupId, memberId, &memberObj).Do()
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(fmt.Sprintf("groups/%s/members/%s", groupId, member.Id))

		// UPDATE will respond with the Group Member that will be created, however, it is eventually consistent
		// After UPDATE, the etag is updated along with the Group Member,
		// once we get a consistent etag, we can feel confident that our Group Member is also consistent
		cc := consistencyCheck{
			resourceType: "group_member",
			timeout:      d.Timeout(schema.TimeoutUpdate),
		}
		err = retryTimeDuration(ctx, d.Timeout(schema.TimeoutUpdate), func() error {
			var retryErr error

			if cc.reachedConsistency(1) {
				return nil
			}

			newMember, retryErr := membersService.Get(groupId, member.Id).IfNoneMatch(cc.lastEtag).Do()
			if googleapi.IsNotModified(retryErr) {
				cc.currConsistent += 1
			} else if retryErr != nil {
				return fmt.Errorf("unexpected error during retries of %s: %s", cc.resourceType, retryErr)
			} else {
				cc.handleNewEtag(newMember.Etag)
			}

			return fmt.Errorf("timed out while waiting for %s to be updated", cc.resourceType)
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	log.Printf("[DEBUG] Finished creating Group Member %q: %#v", memberId, email)

	return resourceGroupMemberRead(ctx, d, meta)
}

func resourceGroupMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	email := d.Get("email").(string)
	groupId := d.Get("group_id").(string)
	memberId := d.Get("member_id").(string)
	log.Printf("[DEBUG] Deleting Group Member %q from Group %s: %#v", memberId, groupId, email)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	membersService, diags := GetMembersService(directoryService)
	if diags.HasError() {
		return diags
	}

	err := membersService.Delete(groupId, memberId).Do()
	if err != nil {
		return handleNotFoundError(err, d, d.Id())
	}

	log.Printf("[DEBUG] Finished deleting Group Member %q: %#v", memberId, email)

	return diags
}

func resourceGroupMemberImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")

	// id is of format "groups/<group_id>/members/<member_id>"
	if len(parts) != 4 {
		return nil, fmt.Errorf("Group Member Id (%s) is not of the correct format (groups/<group_id>/members/<member_id>)", d.Id())
	}

	d.Set("group_id", parts[1])
	d.Set("member_id", parts[3])

	return []*schema.ResourceData{d}, nil
}
