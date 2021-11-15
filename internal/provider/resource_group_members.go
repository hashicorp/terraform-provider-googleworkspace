package googleworkspace

import (
	"context"
	"fmt"
	"google.golang.org/api/googleapi"
	"log"
	"reflect"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	directory "google.golang.org/api/admin/directory/v1"
)

const deliverySettingsDefault = "ALL_MAIL"

type MemberChange struct {
	Old, New map[string]interface{}
}

func resourceGroupMembers() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Group Members resource manages Google Workspace Groups Members. Group Members resides under the " +
			"`https://www.googleapis.com/auth/admin.directory.group` client scope.",

		CreateContext: resourceGroupMembersCreate,
		ReadContext:   resourceGroupMembersRead,
		UpdateContext: resourceGroupMembersUpdate,
		DeleteContext: resourceGroupMembersDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceGroupMembersImport,
		},

		Schema: map[string]*schema.Schema{
			"group_id": {
				Description: "Identifies the group in the API request. The value can be the group's email address, " +
					"group alias, or the unique group ID.",
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"members": {
				Description: "The members of the group",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"email": {
							Description: "The member's email address. A member can be a user or another group. This property is" +
								"required when adding a member to a group. The email must be unique and cannot be an alias of " +
								"another group. If the email address is changed, the API automatically reflects the email address changes.",
							Type:     schema.TypeString,
							Required: true,
						},
						"role": {
							Description: "The member's role in a group. The API returns an error for cycles in group memberships. " +
								"For example, if group1 is a member of group2, group2 cannot be a member of group1. " +
								"Acceptable values are: " +
								"`MANAGER`: This role is only available if the Google Groups for Business is " +
								"enabled using the Admin Console. A `MANAGER` role can do everything done by an `OWNER` role except " +
								"make a member an `OWNER` or delete the group. A group can have multiple `MANAGER` members. " +
								"`MEMBER`: This role can subscribe to a group, view discussion archives, and view the group's " +
								"membership list. " +
								"`OWNER`: This role can send messages to the group, add or remove members, change member roles, " +
								"change group's settings, and delete the group. An OWNER must be a member of the group. " +
								"A group can have more than one OWNER.",
							Type:     schema.TypeString,
							Optional: true,
							Default:  "MEMBER",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"MANAGER", "MEMBER", "OWNER"},
								false)),
						},
						"type": {
							Description: "The type of group member. Acceptable values are: " +
								"`CUSTOMER`: The member represents all users in a domain. An email address is not returned and the " +
								"ID returned is the customer ID. " +
								"`GROUP`: The member is another group. " +
								"`USER`: The member is a user.",
							Type:     schema.TypeString,
							Optional: true,
							Default:  "USER",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"CUSTOMER", "GROUP", "USER"},
								false)),
						},
						"delivery_settings": {
							Description: "Defines mail delivery preferences of member. Acceptable values are:" +
								"`ALL_MAIL`: All messages, delivered as soon as they arrive. " +
								"`DAILY`: No more than one message a day. " +
								"`DIGEST`: Up to 25 messages bundled into a single message. " +
								"`DISABLED`: Remove subscription. " +
								"`NONE`: No messages.",
							Type:     schema.TypeString,
							Optional: true,
							Default:  deliverySettingsDefault,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"ALL_MAIL", "DAILY", "DIGEST",
								"DISABLED", "NONE"}, false)),
						},
						"status": {
							Description: "Status of member.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"id": {
							Description: "The unique ID of the group member. A member id can be used as a member request URI's memberKey.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
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

func resourceGroupMembersCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)
	groupId := d.Get("group_id").(string)

	log.Printf("[DEBUG] Creating Group Members in group %s", groupId)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	membersService, diags := GetMembersService(directoryService)
	if diags.HasError() {
		return diags
	}

	members := d.Get("members").(*schema.Set)
	for _, mMap := range members.List() {
		memb := mMap.(map[string]interface{})

		memberObj := directory.Member{
			Email:            memb["email"].(string),
			Role:             memb["role"].(string),
			Type:             memb["type"].(string),
			DeliverySettings: memb["delivery_settings"].(string),
		}

		log.Printf("[DEBUG] Creating Group Member %q in group %s: %#v", memberObj.Email, groupId, memberObj.Email)

		_, err := membersService.Insert(groupId, &memberObj).Do()
		// If we receive a 409 that the member already exists, ignore it, we'll import it next
		if err != nil && !memberExistsError(err) {
			return diag.FromErr(err)
		}
	}

	d.SetId(fmt.Sprintf("groups/%s", groupId))

	return resourceGroupMembersRead(ctx, d, meta)
}

func resourceGroupMembersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	pageToken := ""
	haveNextPage := true
	groupMembers := make([]*directory.Member, 0)
	for haveNextPage {
		membersObj, err := membersService.List(groupId).PageToken(pageToken).Do()
		if err != nil {
			return handleNotFoundError(err, d, d.Id())
		}
		for _, v := range membersObj.Members {
			groupMembers = append(groupMembers, v)
		}
		pageToken = membersObj.NextPageToken
		if pageToken == "" {
			haveNextPage = false
		}
	}

	configMembers := d.Get("members").(*schema.Set)

	members := make([]interface{}, len(groupMembers))
	for i, member := range groupMembers {

		// Use value if present or default as "delivery_settings" is not provided by API
		deliverySettings := deliverySettingsDefault

		for _, cm := range configMembers.List() {
			cMem := cm.(map[string]interface{})
			if cMem["email"].(string) == member.Email {
				if cMem["delivery_settings"] == "" {
					continue
				}

				deliverySettings = cMem["delivery_settings"].(string)
				break
			}
		}

		members[i] = map[string]interface{}{
			"email":             member.Email,
			"role":              member.Role,
			"type":              member.Type,
			"status":            member.Status,
			"delivery_settings": deliverySettings,
			"id":                member.Id,
		}
	}

	if err := d.Set("members", members); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Error setting attribute",
			Detail:        err.Error(),
			AttributePath: cty.IndexStringPath("members"),
		})
	}

	d.SetId(fmt.Sprintf("groups/%s", groupId))

	return diags
}

func resourceGroupMembersUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	groupId := d.Get("group_id").(string)
	log.Printf("[DEBUG] Updating Group Members of group: %s", groupId)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	membersService, diags := GetMembersService(directoryService)
	if diags.HasError() {
		return diags
	}

	o, n := d.GetChange("members")
	vals := make(map[string]*MemberChange)
	for _, raw := range o.(*schema.Set).List() {
		obj := raw.(map[string]interface{})
		k := obj["email"].(string)
		vals[k] = &MemberChange{Old: obj}
	}
	for _, raw := range n.(*schema.Set).List() {
		obj := raw.(map[string]interface{})
		k := obj["email"].(string)
		if _, ok := vals[k]; !ok {
			vals[k] = &MemberChange{}
		}
		vals[k].New = obj
	}

	for name, change := range vals {
		// Create a new one if old is nil
		if change.Old == nil {
			memberObj := directory.Member{
				Email:            change.New["email"].(string),
				Role:             change.New["role"].(string),
				Type:             change.New["type"].(string),
				DeliverySettings: change.New["delivery_settings"].(string),
			}

			log.Printf("[DEBUG] Creating Group Member %q in group %s: %#v", memberObj.Email, groupId, memberObj.Email)

			_, err := membersService.Insert(groupId, &memberObj).Do()
			if err != nil {
				return diag.FromErr(err)
			}
			continue
		}
		// Delete member if new is nil
		if change.New == nil {
			memberKey := change.Old["id"].(string)
			log.Printf("[DEBUG] Remove Group Member %q from group %s: %#v", name, groupId, memberKey)
			err := membersService.Delete(groupId, memberKey).Do()
			if err != nil {
				return diag.FromErr(err)
			}
			continue
		}
		// no change
		if reflect.DeepEqual(change.Old, change.New) {
			continue
		}

		memberObj := directory.Member{
			Email:            change.New["email"].(string),
			Role:             change.New["role"].(string),
			Type:             change.New["type"].(string),
			DeliverySettings: change.New["delivery_settings"].(string),
		}

		_, err := membersService.Update(groupId, change.Old["id"].(string), &memberObj).Do()
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(fmt.Sprintf("groups/%s", groupId))
		log.Printf("[DEBUG] Finished updating Group Members %q", groupId)
	}

	return resourceGroupMembersRead(ctx, d, meta)
}

func resourceGroupMembersDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	groupId := d.Get("group_id").(string)
	members := d.Get("members").(*schema.Set)
	log.Printf("[DEBUG] Deleting Group Members from Group %s", groupId)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	membersService, diags := GetMembersService(directoryService)
	if diags.HasError() {
		return diags
	}

	for _, raw := range members.List() {
		member := raw.(map[string]interface{})
		memberKey := member["id"].(string)
		err := membersService.Delete(groupId, memberKey).Do()
		if err != nil {
			return handleNotFoundError(err, d, d.Id())
		}
		log.Printf("[DEBUG] Finished deleting Group Member %q: %#v", memberKey, member["email"].(string))
	}

	log.Printf("[DEBUG] Finished deleting Group Members %s", groupId)

	return diags
}

func resourceGroupMembersImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")

	// id is of format "groups/<group_id>"
	if len(parts) != 2 {
		return nil, fmt.Errorf("Group Member Id (%s) is not of the correct format (groups/<group_id>)", d.Id())
	}

	d.Set("group_id", parts[1])

	return []*schema.ResourceData{d}, nil
}

func memberExistsError(err error) bool {
	gerr, ok := err.(*googleapi.Error)
	if !ok {
		return false
	}

	if gerr.Code == 409 && strings.Contains(gerr.Body, "Member already exists") {
		log.Printf("[DEBUG] Dismissed an error based on error code: %s", err)
		return true
	}
	return false
}
