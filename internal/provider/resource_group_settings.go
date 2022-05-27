package googleworkspace

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/groupssettings/v1"
)

func resourceGroupSettings() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Group Settings resource manages Google Workspace Groups Setting. Group Settings requires the " +
			"`https://www.googleapis.com/auth/apps.groups.settings` client scope.",

		CreateContext: resourceGroupSettingsCreate,
		ReadContext:   resourceGroupSettingsRead,
		UpdateContext: resourceGroupSettingsUpdate,
		DeleteContext: resourceGroupSettingsDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"email": {
				Description: "The group's email address.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "Name of the group, which has a maximum size of 75 characters.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "Description of the group. The maximum group description is no more than 300 characters.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"who_can_join": {
				Description: "Permission to join group. Possible values are: " +
					"\n\t- `ANYONE_CAN_JOIN`: Any Internet user, both inside and outside your domain, can join the group. " +
					"\n\t- `ALL_IN_DOMAIN_CAN_JOIN`: Anyone in the account domain can join. This includes accounts with multiple domains. " +
					"\n\t- `INVITED_CAN_JOIN`: Candidates for membership can be invited to join. " +
					"\n\t- `CAN_REQUEST_TO_JOIN`: Non members can request an invitation to join.",
				Type:     schema.TypeString,
				Optional: true,
				Default:  "CAN_REQUEST_TO_JOIN",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"ANYONE_CAN_JOIN",
					"ALL_IN_DOMAIN_CAN_JOIN", "INVITED_CAN_JOIN", "CAN_REQUEST_TO_JOIN"}, true)),
			},
			"who_can_view_membership": {
				Description: "Permissions to view membership. Possible values are: " +
					"\n\t- `ALL_IN_DOMAIN_CAN_VIEW`: Anyone in the account can view the group members list. " +
					"If a group already has external members, those members can still send email to this group. " +
					"\n\t- `ALL_MEMBERS_CAN_VIEW`: The group members can view the group members list. " +
					"\n\t- `ALL_MANAGERS_CAN_VIEW`: The group managers can view group members list. " +
					"\n\t- `ALL_OWNERS_CAN_VIEW`: The group owners can view group members list.",
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ALL_MEMBERS_CAN_VIEW",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"ALL_IN_DOMAIN_CAN_VIEW",
					"ALL_MEMBERS_CAN_VIEW", "ALL_MANAGERS_CAN_VIEW", "ALL_OWNERS_CAN_VIEW"}, true)),
			},
			"who_can_view_group": {
				Description: "Permissions to view group messages. Possible values are: " +
					"\n\t- `ANYONE_CAN_VIEW`: Any Internet user can view the group's messages. " +
					"\n\t- `ALL_IN_DOMAIN_CAN_VIEW`: Anyone in your account can view this group's messages. " +
					"\n\t- `ALL_MEMBERS_CAN_VIEW`: All group members can view the group's messages. " +
					"\n\t- `ALL_MANAGERS_CAN_VIEW`: Any group manager can view this group's messages. " +
					"\n\t- `ALL_OWNERS_CAN_VIEW`: The group owners can view this group's messages.",
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ALL_MEMBERS_CAN_VIEW",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"ANYONE_CAN_VIEW",
					"ALL_IN_DOMAIN_CAN_VIEW", "ALL_MEMBERS_CAN_VIEW", "ALL_MANAGERS_CAN_VIEW", "ALL_OWNERS_CAN_VIEW"},
					true)),
			},
			"allow_external_members": {
				Description: "Identifies whether members external to your organization can join the group. If true, " +
					"Google Workspace users external to your organization can become members of this group. If false, " +
					"users not belonging to the organization are not allowed to become members of this group.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"who_can_post_message": {
				Description: "Permissions to post messages. Possible values are: " +
					"\n\t- `NONE_CAN_POST`: The group is disabled and archived. No one can post a message to this group. " +
					"* When archiveOnly is false, updating whoCanPostMessage to NONE_CAN_POST, results in an error. " +
					"* If archiveOnly is reverted from true to false, whoCanPostMessages is set to ALL_MANAGERS_CAN_POST. " +
					"\n\t- `ALL_MANAGERS_CAN_POST`: Managers, including group owners, can post messages. " +
					"\n\t- `ALL_MEMBERS_CAN_POST`: Any group member can post a message. " +
					"\n\t- `ALL_OWNERS_CAN_POST`: Only group owners can post a message. " +
					"\n\t- `ALL_IN_DOMAIN_CAN_POST`: Anyone in the account can post a message. " +
					"\n\t- `ANYONE_CAN_POST`: Any Internet user who outside your account can access your Google Groups " +
					"service and post a message. " +
					"\n\t" +
					"*Note: When `who_can_post_message` is set to `ANYONE_CAN_POST`, we recommend the" +
					"`message_moderation_level` be set to `MODERATE_NON_MEMBERS` to protect the group from possible spam. " +
					"Users not belonging to the organization are not allowed to become members of this group.",
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"NONE_CAN_POST",
					"ALL_MANAGERS_CAN_POST", "ALL_MEMBERS_CAN_POST", "ALL_OWNERS_CAN_POST", "ALL_IN_DOMAIN_CAN_POST",
					"ANYONE_CAN_POST"}, true)),
			},
			"allow_web_posting": {
				Description: "Allows posting from web. If true, allows any member to post to the group forum. If false, " +
					"Members only use Gmail to communicate with the group.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"primary_language": {
				Description: "The primary language for group. For a group's primary language use the language tags from " +
					"the Google Workspace languages found at Google Workspace Email Settings API Email Language Tags.",
				Type:     schema.TypeString,
				Optional: true,
			},
			"is_archived": {
				Description: "Allows the Group contents to be archived. If true, archive messages sent to the group. " +
					"If false, Do not keep an archive of messages sent to this group. If false, previously archived " +
					"messages remain in the archive.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"archive_only": {
				Description: "Allows the group to be archived only. If true, Group is archived and the group is inactive. " +
					"New messages to this group are rejected. The older archived messages are browsable and searchable. " +
					"If true, the `who_can_post_message` property is set to `NONE_CAN_POST`. If reverted from true to false, " +
					"`who_can_post_message` is set to `ALL_MANAGERS_CAN_POST`. If false, The group is active and can " +
					"receive messages. When false, updating `who_can_post_message` to `NONE_CAN_POST`, results in an error.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"message_moderation_level": {
				Description: "Moderation level of incoming messages. Possible values are: " +
					"\n\t- `MODERATE_ALL_MESSAGES`: All messages are sent to the group owner's email address for approval. " +
					"If approved, the message is sent to the group. " +
					"\n\t- `MODERATE_NON_MEMBERS`: All messages from non group members are sent to the group owner's email " +
					"address for approval. If approved, the message is sent to the group. " +
					"\n\t- `MODERATE_NEW_MEMBERS`: All messages from new members are sent to the group owner's email address " +
					"for approval. If approved, the message is sent to the group. " +
					"\n\t- `MODERATE_NONE`: No moderator approval is required. Messages are delivered directly to the group." +
					"\n\t" +
					"Note: When the `who_can_post_message` is set to `ANYONE_CAN_POST`, we recommend the " +
					"`message_moderation_level` be set to `MODERATE_NON_MEMBERS` to protect the group from possible spam." +
					"When `member_can_post_as_the_group` is true, any message moderation settings on individual users " +
					"or new members will not apply to posts made on behalf of the group.",
				Type:     schema.TypeString,
				Optional: true,
				Default:  "MODERATE_NONE",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"MODERATE_ALL_MESSAGES",
					"MODERATE_NON_MEMBERS", "MODERATE_NEW_MEMBERS", "MODERATE_NONE"}, true)),
			},
			"spam_moderation_level": {
				Description: "Specifies moderation levels for messages detected as spam. Possible values are: " +
					"\n\t- `ALLOW`: Post the message to the group. " +
					"\n\t- `MODERATE`: Send the message to the moderation queue. This is the default. " +
					"\n\t- `SILENTLY_MODERATE`: Send the message to the moderation queue, but do not send notification to moderators. " +
					"\n\t- `REJECT`: Immediately reject the message.",
				Type:     schema.TypeString,
				Optional: true,
				Default:  "MODERATE",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"ALLOW",
					"MODERATE", "SILENTLY_MODERATE", "REJECT"}, true)),
			},
			"reply_to": {
				Description: "Specifies who receives the default reply. Possible values are: " +
					"\n\t- `REPLY_TO_CUSTOM`: For replies to messages, use the group's custom email address. " +
					"When set to `REPLY_TO_CUSTOM`, the `custom_reply_to` property holds the custom email address used " +
					"when replying to a message, the customReplyTo property must have a value. Otherwise an error is returned. " +
					"\n\t- `REPLY_TO_SENDER`: The reply sent to author of message. " +
					"\n\t- `REPLY_TO_LIST`: This reply message is sent to the group. " +
					"\n\t- `REPLY_TO_OWNER`: The reply is sent to the owner(s) of the group. This does not include the group's managers. " +
					"\n\t- `REPLY_TO_IGNORE`: Group users individually decide where the message reply is sent. " +
					"\n\t- `REPLY_TO_MANAGERS`: This reply message is sent to the group's managers, which includes all " +
					"managers and the group owner.",
				Type:     schema.TypeString,
				Optional: true,
				Default:  "REPLY_TO_IGNORE",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"REPLY_TO_CUSTOM",
					"REPLY_TO_SENDER", "REPLY_TO_LIST", "REPLY_TO_OWNER", "REPLY_TO_IGNORE",
					"REPLY_TO_MANAGERS"}, true)),
			},
			"custom_reply_to": {
				Description: "An email address used when replying to a message if the `reply_to` property is set to " +
					"`REPLY_TO_CUSTOM`. This address is defined by an account administrator. When the group's `reply_to` " +
					"property is set to `REPLY_TO_CUSTOM`, the `custom_reply_to` property holds a custom email address " +
					"used when replying to a message, the `custom_reply_to` property must have a text value or an error is returned.",
				Type:     schema.TypeString,
				Optional: true,
			},
			"include_custom_footer": {
				Description: "Whether to include custom footer.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"custom_footer_text": {
				Description:      "Set the content of custom footer text. The maximum number of characters is 1,000.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(0, 1000)),
			},
			"send_message_deny_notification": {
				Description: "Allows a member to be notified if the member's message to the group is denied by the " +
					"group owner. If true, when a message is rejected, send the deny message notification to the " +
					"message author. The `default_message_deny_notification_text` property is dependent on the " +
					"`send_message_deny_notification` property being true. If false, when a message is rejected, " +
					"no notification is sent.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"default_message_deny_notification_text": {
				Description: "When a message is rejected, this is text for the rejection notification sent to the " +
					"message's author. By default, this property is empty and has no value in the API's response body. " +
					"The maximum notification text size is 10,000 characters. Requires `send_message_deny_notification` " +
					"property to be true.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(0, 10000)),
			},
			"members_can_post_as_the_group": {
				Description: "Enables members to post messages as the group. If true, group member can post messages " +
					"using the group's email address instead of their own email address. Message appear to originate " +
					"from the group itself. Any message moderation settings on individual users or new members do not " +
					"apply to posts made on behalf of the group. If false, members can not post in behalf of the " +
					"group's email address.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"include_in_global_address_list": {
				Description: "Enables the group to be included in the Global Address List. If true, the group is " +
					"included in the Global Address List. If false, it is not included in the Global Address List.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"who_can_leave_group": {
				Description: "Permission to leave the group. Possible values are:" +
					"\n\t- `ALL_MANAGERS_CAN_LEAVE`" +
					"\n\t- `ALL_MEMBERS_CAN_LEAVE`" +
					"\n\t- `NONE_CAN_LEAVE`",
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ALL_MEMBERS_CAN_LEAVE",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"ALL_MANAGERS_CAN_LEAVE",
					"ALL_MEMBERS_CAN_LEAVE", "NONE_CAN_LEAVE"}, true)),
			},
			"who_can_contact_owner": {
				Description: "Permission to contact owner of the group via web UI. Possible values are: " +
					"\n\t- `ALL_IN_DOMAIN_CAN_CONTACT`" +
					"\n\t- `ALL_MANAGERS_CAN_CONTACT`" +
					"\n\t- `ALL_MEMBERS_CAN_CONTACT`" +
					"\n\t- `ANYONE_CAN_CONTACT`" +
					"\n\t- `ALL_OWNERS_CAN_CONTACT`",
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ANYONE_CAN_CONTACT",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"ALL_IN_DOMAIN_CAN_CONTACT",
					"ALL_MANAGERS_CAN_CONTACT", "ALL_MEMBERS_CAN_CONTACT", "ANYONE_CAN_CONTACT", "ALL_OWNERS_CAN_CONTACT"}, true)),
			},
			"who_can_moderate_members": {
				Description: "Specifies who can manage members. Possible values are: " +
					"\n\t- `ALL_MEMBERS`" +
					"\n\t- `OWNERS_AND_MANAGERS`" +
					"\n\t- `OWNERS_ONLY`" +
					"\n\t- `NONE`",
				Type:     schema.TypeString,
				Optional: true,
				Default:  "OWNERS_AND_MANAGERS",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"ALL_MEMBERS",
					"OWNERS_AND_MANAGERS", "OWNERS_ONLY", "NONE"}, true)),
			},
			"who_can_moderate_content": {
				Description: "Specifies who can moderate content. Possible values are: " +
					"\n\t- `ALL_MEMBERS`" +
					"\n\t- `OWNERS_AND_MANAGERS`" +
					"\n\t- `OWNERS_ONLY`" +
					"\n\t- `NONE`",
				Type:     schema.TypeString,
				Optional: true,
				Default:  "OWNERS_AND_MANAGERS",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"ALL_MEMBERS",
					"OWNERS_AND_MANAGERS", "OWNERS_ONLY", "NONE"}, true)),
			},
			"who_can_assist_content": {
				Description: "Specifies who can moderate metadata. Possible values are: " +
					"\n\t- `ALL_MEMBERS`" +
					"\n\t- `OWNERS_AND_MANAGERS`" +
					"\n\t- `MANAGERS_ONLY`" +
					"\n\t- `OWNERS_ONLY`" +
					"\n\t- `NONE`",
				Type:     schema.TypeString,
				Optional: true,
				Default:  "NONE",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"ALL_MEMBERS",
					"OWNERS_AND_MANAGERS", "MANAGERS_ONLY", "OWNERS_ONLY", "NONE"}, true)),
			},
			"custom_roles_enabled_for_settings_to_be_merged": {
				Description: "Specifies whether the group has a custom role that's included in one of the settings " +
					"being merged.",
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enable_collaborative_inbox": {
				Description: "Specifies whether a collaborative inbox will remain turned on for the group.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"who_can_discover_group": {
				Description: "Specifies the set of users for whom this group is discoverable. Possible values are: " +
					"\n\t- `ANYONE_CAN_DISCOVER`" +
					"\n\t- `ALL_IN_DOMAIN_CAN_DISCOVER`" +
					"\n\t- `ALL_MEMBERS_CAN_DISCOVER`",
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ALL_IN_DOMAIN_CAN_DISCOVER",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"ANYONE_CAN_DISCOVER",
					"ALL_IN_DOMAIN_CAN_DISCOVER", "ALL_MEMBERS_CAN_DISCOVER"}, true)),
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

func resourceGroupSettingsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	email := d.Get("email").(string)
	log.Printf("[DEBUG] Creating Group Settings %q: %#v", email, email)

	groupsSettingsService, diags := client.NewGroupsSettingsService()
	if diags.HasError() {
		return diags
	}

	groupsService, diags := GetGroupsSettingsService(groupsSettingsService)
	if diags.HasError() {
		return diags
	}

	groupSettingsObj := groupssettings.Groups{
		Email:                                   email,
		Name:                                    d.Get("name").(string),
		Description:                             d.Get("description").(string),
		WhoCanJoin:                              d.Get("who_can_join").(string),
		WhoCanViewMembership:                    d.Get("who_can_view_membership").(string),
		WhoCanViewGroup:                         d.Get("who_can_view_group").(string),
		AllowExternalMembers:                    strconv.FormatBool(d.Get("allow_external_members").(bool)),
		WhoCanPostMessage:                       d.Get("who_can_post_message").(string),
		AllowWebPosting:                         strconv.FormatBool(d.Get("allow_web_posting").(bool)),
		PrimaryLanguage:                         d.Get("primary_language").(string),
		IsArchived:                              strconv.FormatBool(d.Get("is_archived").(bool)),
		ArchiveOnly:                             strconv.FormatBool(d.Get("archive_only").(bool)),
		MessageModerationLevel:                  d.Get("message_moderation_level").(string),
		SpamModerationLevel:                     d.Get("spam_moderation_level").(string),
		ReplyTo:                                 d.Get("reply_to").(string),
		CustomReplyTo:                           d.Get("custom_reply_to").(string),
		IncludeCustomFooter:                     strconv.FormatBool(d.Get("include_custom_footer").(bool)),
		CustomFooterText:                        d.Get("custom_footer_text").(string),
		SendMessageDenyNotification:             strconv.FormatBool(d.Get("send_message_deny_notification").(bool)),
		DefaultMessageDenyNotificationText:      d.Get("default_message_deny_notification_text").(string),
		MembersCanPostAsTheGroup:                strconv.FormatBool(d.Get("members_can_post_as_the_group").(bool)),
		IncludeInGlobalAddressList:              strconv.FormatBool(d.Get("include_in_global_address_list").(bool)),
		WhoCanLeaveGroup:                        d.Get("who_can_leave_group").(string),
		WhoCanContactOwner:                      d.Get("who_can_contact_owner").(string),
		WhoCanModerateMembers:                   d.Get("who_can_moderate_members").(string),
		WhoCanModerateContent:                   d.Get("who_can_moderate_content").(string),
		WhoCanAssistContent:                     d.Get("who_can_assist_content").(string),
		CustomRolesEnabledForSettingsToBeMerged: strconv.FormatBool(d.Get("custom_roles_enabled_for_settings_to_be_merged").(bool)),
		EnableCollaborativeInbox:                strconv.FormatBool(d.Get("enable_collaborative_inbox").(bool)),
		WhoCanDiscoverGroup:                     d.Get("who_can_discover_group").(string),

		ForceSendFields: []string{"AllowExternalMembers", "AllowWebPosting", "IsArchived", "ArchiveOnly",
			"IncludeCustomFooter", "SendMessageDenyNotification", "MembersCanPostAsTheGroup", "IncludeInGlobalAddressList",
			"CustomRolesEnabledForSettingsToBeMerged", "EnableCollaborativeInbox"},
	}

	groupSettings, err := groupsService.Update(email, &groupSettingsObj).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(groupSettings.Email)

	numInserts := 1
	cc := consistencyCheck{
		timeout:      d.Timeout(schema.TimeoutCreate),
		resourceType: "group_settings",
	}
	err = retryTimeDuration(ctx, d.Timeout(schema.TimeoutCreate), func() error {
		if cc.reachedConsistency(numInserts) {
			return nil
		}

		newGroupSettings, retryErr := groupsService.Get(d.Id()).IfNoneMatch(cc.lastEtag).Do()
		if googleapi.IsNotModified(retryErr) {
			cc.currConsistent += 1
		} else if retryErr != nil {
			return fmt.Errorf("unexpected error during retries of %s: %s", cc.resourceType, retryErr)
		} else {
			cc.handleNewEtag(newGroupSettings.ServerResponse.Header.Get("Etag"))
		}

		return fmt.Errorf("timed out while waiting for group settings to be updated")
	})

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished creating Group Settings %q: %#v", d.Id(), email)

	return resourceGroupSettingsRead(ctx, d, meta)
}

func resourceGroupSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	groupsSettingsService, diags := client.NewGroupsSettingsService()
	if diags.HasError() {
		return diags
	}

	groupsService, diags := GetGroupsSettingsService(groupsSettingsService)
	if diags.HasError() {
		return diags
	}

	group, err := groupsService.Get(d.Id()).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	// Convert strings to bools
	allowExternalMembers, err := strconv.ParseBool(group.AllowExternalMembers)
	if err != nil {
		return diag.FromErr(err)
	}

	allowWebPosting, err := strconv.ParseBool(group.AllowWebPosting)
	if err != nil {
		return diag.FromErr(err)
	}

	isArchived, err := strconv.ParseBool(group.IsArchived)
	if err != nil {
		return diag.FromErr(err)
	}

	archiveOnly, err := strconv.ParseBool(group.ArchiveOnly)
	if err != nil {
		return diag.FromErr(err)
	}

	includeCustomFooter, err := strconv.ParseBool(group.IncludeCustomFooter)
	if err != nil {
		return diag.FromErr(err)
	}

	sendMessageDenyNotification, err := strconv.ParseBool(group.SendMessageDenyNotification)
	if err != nil {
		return diag.FromErr(err)
	}

	membersCanPostAsTheGroup, err := strconv.ParseBool(group.MembersCanPostAsTheGroup)
	if err != nil {
		return diag.FromErr(err)
	}

	includeInGlobalAddressList, err := strconv.ParseBool(group.IncludeInGlobalAddressList)
	if err != nil {
		return diag.FromErr(err)
	}

	customRolesEnabledForSettingsToBeMerged, err := strconv.ParseBool(group.CustomRolesEnabledForSettingsToBeMerged)
	if err != nil {
		return diag.FromErr(err)
	}

	enableCollaborativeInbox, err := strconv.ParseBool(group.EnableCollaborativeInbox)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("email", group.Email)
	d.Set("name", group.Name)
	d.Set("description", group.Description)
	d.Set("who_can_join", group.WhoCanJoin)
	d.Set("who_can_view_membership", group.WhoCanViewMembership)
	d.Set("who_can_view_group", group.WhoCanViewGroup)
	d.Set("allow_external_members", allowExternalMembers)
	d.Set("who_can_post_message", group.WhoCanPostMessage)
	d.Set("allow_web_posting", allowWebPosting)
	d.Set("primary_language", group.PrimaryLanguage)
	d.Set("is_archived", isArchived)
	d.Set("archive_only", archiveOnly)
	d.Set("message_moderation_level", group.MessageModerationLevel)
	d.Set("spam_moderation_level", group.SpamModerationLevel)
	d.Set("reply_to", group.ReplyTo)
	d.Set("custom_reply_to", group.CustomReplyTo)
	d.Set("include_custom_footer", includeCustomFooter)
	d.Set("custom_footer_text", group.CustomFooterText)
	d.Set("send_message_deny_notification", sendMessageDenyNotification)
	d.Set("default_message_deny_notification_text", group.DefaultMessageDenyNotificationText)
	d.Set("members_can_post_as_the_group", membersCanPostAsTheGroup)
	d.Set("include_in_global_address_list", includeInGlobalAddressList)
	d.Set("who_can_leave_group", group.WhoCanLeaveGroup)
	d.Set("who_can_contact_owner", group.WhoCanContactOwner)
	d.Set("who_can_moderate_members", group.WhoCanModerateMembers)
	d.Set("who_can_moderate_content", group.WhoCanModerateContent)
	d.Set("who_can_assist_content", group.WhoCanAssistContent)
	d.Set("custom_roles_enabled_for_settings_to_be_merged", customRolesEnabledForSettingsToBeMerged)
	d.Set("enable_collaborative_inbox", enableCollaborativeInbox)
	d.Set("who_can_discover_group", group.WhoCanDiscoverGroup)

	d.SetId(group.Email)

	return diags
}

func resourceGroupSettingsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	email := d.Get("email").(string)
	log.Printf("[DEBUG] Updating Group Settings %q: %#v", email, email)

	groupsSettingsService, diags := client.NewGroupsSettingsService()
	if diags.HasError() {
		return diags
	}

	groupsService, diags := GetGroupsSettingsService(groupsSettingsService)
	if diags.HasError() {
		return diags
	}

	groupSettingsObj := groupssettings.Groups{}

	forceSendFields := []string{}

	if d.HasChange("who_can_join") {
		groupSettingsObj.WhoCanJoin = d.Get("who_can_join").(string)
	}

	if d.HasChange("who_can_view_membership") {
		groupSettingsObj.WhoCanViewMembership = d.Get("who_can_view_membership").(string)
	}

	if d.HasChange("who_can_view_group") {
		groupSettingsObj.WhoCanViewGroup = d.Get("who_can_view_group").(string)
	}

	if d.HasChange("allow_external_members") {
		groupSettingsObj.AllowExternalMembers = strconv.FormatBool(d.Get("allow_external_members").(bool))
		forceSendFields = append(forceSendFields, "AllowExternalMembers")
	}

	if d.HasChange("who_can_post_message") {
		groupSettingsObj.WhoCanPostMessage = d.Get("who_can_post_message").(string)
	}

	if d.HasChange("allow_web_posting") {
		groupSettingsObj.AllowWebPosting = strconv.FormatBool(d.Get("allow_web_posting").(bool))
		forceSendFields = append(forceSendFields, "AllowWebPosting")
	}

	if d.HasChange("primary_language") {
		groupSettingsObj.PrimaryLanguage = d.Get("primary_language").(string)
	}

	if d.HasChange("is_archived") {
		groupSettingsObj.IsArchived = strconv.FormatBool(d.Get("is_archived").(bool))
		forceSendFields = append(forceSendFields, "IsArchived")
	}

	if d.HasChange("archive_only") {
		groupSettingsObj.ArchiveOnly = strconv.FormatBool(d.Get("archive_only").(bool))
		forceSendFields = append(forceSendFields, "ArchiveOnly")
	}

	if d.HasChange("message_moderation_level") {
		groupSettingsObj.MessageModerationLevel = d.Get("message_moderation_level").(string)
	}

	if d.HasChange("spam_moderation_level") {
		groupSettingsObj.SpamModerationLevel = d.Get("spam_moderation_level").(string)
	}

	if d.HasChange("reply_to") {
		groupSettingsObj.ReplyTo = d.Get("reply_to").(string)
	}

	if d.HasChange("custom_reply_to") {
		groupSettingsObj.CustomReplyTo = d.Get("custom_reply_to").(string)
	}

	if d.HasChange("include_custom_footer") {
		groupSettingsObj.IncludeCustomFooter = strconv.FormatBool(d.Get("include_custom_footer").(bool))
		forceSendFields = append(forceSendFields, "IncludeCustomFooter")
	}

	if d.HasChange("custom_footer_text") {
		groupSettingsObj.CustomFooterText = d.Get("custom_footer_text").(string)
	}

	if d.HasChange("send_message_deny_notification") {
		groupSettingsObj.SendMessageDenyNotification = strconv.FormatBool(d.Get("send_message_deny_notification").(bool))
		forceSendFields = append(forceSendFields, "SendMessageDenyNotification")
	}

	if d.HasChange("default_message_deny_notification_text") {
		groupSettingsObj.DefaultMessageDenyNotificationText = d.Get("default_message_deny_notification_text").(string)
	}

	if d.HasChange("members_can_post_as_the_group") {
		groupSettingsObj.MembersCanPostAsTheGroup = strconv.FormatBool(d.Get("members_can_post_as_the_group").(bool))
		forceSendFields = append(forceSendFields, "MembersCanPostAsTheGroup")
	}

	if d.HasChange("include_in_global_address_list") {
		groupSettingsObj.IncludeInGlobalAddressList = strconv.FormatBool(d.Get("include_in_global_address_list").(bool))
		forceSendFields = append(forceSendFields, "IncludeInGlobalAddressList")
	}

	if d.HasChange("who_can_leave_group") {
		groupSettingsObj.WhoCanLeaveGroup = d.Get("who_can_leave_group").(string)
	}

	if d.HasChange("who_can_contact_owner") {
		groupSettingsObj.WhoCanContactOwner = d.Get("who_can_contact_owner").(string)
	}

	if d.HasChange("who_can_moderate_members") {
		groupSettingsObj.WhoCanModerateMembers = d.Get("who_can_moderate_members").(string)
	}

	if d.HasChange("who_can_moderate_content") {
		groupSettingsObj.WhoCanModerateContent = d.Get("who_can_moderate_content").(string)
	}

	if d.HasChange("who_can_assist_content") {
		groupSettingsObj.WhoCanAssistContent = d.Get("who_can_assist_content").(string)
	}

	if d.HasChange("enable_collaborative_inbox") {
		groupSettingsObj.EnableCollaborativeInbox = strconv.FormatBool(d.Get("enable_collaborative_inbox").(bool))
		forceSendFields = append(forceSendFields, "EnableCollaborativeInbox")
	}

	if d.HasChange("who_can_discover_group") {
		groupSettingsObj.WhoCanDiscoverGroup = d.Get("who_can_discover_group").(string)
	}

	if len(forceSendFields) > 0 {
		groupSettingsObj.ForceSendFields = forceSendFields
	}

	groupSettings, err := groupsService.Update(email, &groupSettingsObj).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(groupSettings.Email)

	numInserts := 1
	cc := consistencyCheck{
		timeout:      d.Timeout(schema.TimeoutUpdate),
		resourceType: "group_settings",
	}
	err = retryTimeDuration(ctx, d.Timeout(schema.TimeoutUpdate), func() error {
		if cc.reachedConsistency(numInserts) {
			return nil
		}

		newGroupSettings, retryErr := groupsService.Get(d.Id()).IfNoneMatch(cc.lastEtag).Do()
		if googleapi.IsNotModified(retryErr) {
			cc.currConsistent += 1
		} else if retryErr != nil {
			return fmt.Errorf("unexpected error during retries of %s: %s", cc.resourceType, retryErr)
		} else {
			cc.handleNewEtag(newGroupSettings.ServerResponse.Header.Get("Etag"))
		}

		return fmt.Errorf("timed out while waiting for group settings to be updated")
	})

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished updating Group Settings %q: %#v", d.Id(), email)

	return resourceGroupSettingsRead(ctx, d, meta)
}

func resourceGroupSettingsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	log.Printf("[DEBUG] Removing Group Settings from state for %q", d.Id())

	d.SetId("")

	return nil
}
