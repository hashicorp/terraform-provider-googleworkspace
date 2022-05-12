package googleworkspace

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-googleworkspace-pf/internal/model"
	"log"
)

type resourceGroupSettingsType struct{}

// GetSchema Group Settings Resource
func (r resourceGroupSettingsType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Group Settings resource manages Google Workspace Groups Setting. Group Settings requires the " +
			"`https://www.googleapis.com/auth/apps.groups.settings` client scope.",
		Attributes: map[string]tfsdk.Attribute{
			"email": {
				Description: "The group's email address.",
				Type:        types.StringType,
				Required:    true,
			},
			"name": {
				Description: "Name of the group, which has a maximum size of 75 characters.",
				Type:        types.StringType,
				Computed:    true,
			},
			"description": {
				Description: "Description of the group. The maximum group description is no more than 300 characters.",
				Type:        types.StringType,
				Computed:    true,
			},
			"who_can_join": {
				Description: "Permission to join group. Possible values are: " +
					"`ANYONE_CAN_JOIN`: Any Internet user, both inside and outside your domain, can join the group. " +
					"`ALL_IN_DOMAIN_CAN_JOIN`: Anyone in the account domain can join. This includes accounts with multiple domains. " +
					"`INVITED_CAN_JOIN`: Candidates for membership can be invited to join. " +
					"`CAN_REQUEST_TO_JOIN`: Non members can request an invitation to join.",
				Type:     types.StringType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.String{Value: "CAN_REQUEST_TO_JOIN"},
					},
				},
				Validators: []tfsdk.AttributeValidator{
					StringInSliceValidator{
						Options: []string{"ANYONE_CAN_JOIN",
							"ALL_IN_DOMAIN_CAN_JOIN", "INVITED_CAN_JOIN", "CAN_REQUEST_TO_JOIN"},
					},
				},
			},
			"who_can_view_membership": {
				Description: "Permissions to view membership. Possible values are: " +
					"`ALL_IN_DOMAIN_CAN_VIEW`: Anyone in the account can view the group members list. " +
					"If a group already has external members, those members can still send email to this group. " +
					"`ALL_MEMBERS_CAN_VIEW`: The group members can view the group members list. " +
					"`ALL_MANAGERS_CAN_VIEW`: The group managers can view group members list. " +
					"`ALL_OWNERS_CAN_VIEW`: The group owners can view group members list.",
				Type:     types.StringType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.String{Value: "ALL_MEMBERS_CAN_VIEW"},
					},
				},
				Validators: []tfsdk.AttributeValidator{
					StringInSliceValidator{
						Options: []string{"ALL_IN_DOMAIN_CAN_VIEW",
							"ALL_MEMBERS_CAN_VIEW", "ALL_MANAGERS_CAN_VIEW", "ALL_OWNERS_CAN_VIEW"},
					},
				},
			},
			"who_can_view_group": {
				Description: "Permissions to view group messages. Possible values are: " +
					"`ANYONE_CAN_VIEW`: Any Internet user can view the group's messages. " +
					"`ALL_IN_DOMAIN_CAN_VIEW`: Anyone in your account can view this group's messages. " +
					"`ALL_MEMBERS_CAN_VIEW`: All group members can view the group's messages. " +
					"`ALL_MANAGERS_CAN_VIEW`: Any group manager can view this group's messages. " +
					"`ALL_OWNERS_CAN_VIEW`: The group owners can view this group's messages.",
				Type:     types.StringType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.String{Value: "ALL_MEMBERS_CAN_VIEW"},
					},
				},
				Validators: []tfsdk.AttributeValidator{
					StringInSliceValidator{
						Options: []string{"ANYONE_CAN_VIEW",
							"ALL_IN_DOMAIN_CAN_VIEW", "ALL_MEMBERS_CAN_VIEW", "ALL_MANAGERS_CAN_VIEW", "ALL_OWNERS_CAN_VIEW"},
					},
				},
			},
			"allow_external_members": {
				Description: "Identifies whether members external to your organization can join the group. If true, " +
					"Google Workspace users external to your organization can become members of this group. If false, " +
					"users not belonging to the organization are not allowed to become members of this group.",
				Type:     types.BoolType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.Bool{Value: false},
					},
				},
			},
			"who_can_post_message": {
				Description: "Permissions to post messages. Possible values are: " +
					"`NONE_CAN_POST`: The group is disabled and archived. No one can post a message to this group. " +
					"* When archiveOnly is false, updating whoCanPostMessage to NONE_CAN_POST, results in an error. " +
					"* If archiveOnly is reverted from true to false, whoCanPostMessages is set to ALL_MANAGERS_CAN_POST. " +
					"`ALL_MANAGERS_CAN_POST`: Managers, including group owners, can post messages. " +
					"`ALL_MEMBERS_CAN_POST`: Any group member can post a message. " +
					"`ALL_OWNERS_CAN_POST`: Only group owners can post a message. " +
					"`ALL_IN_DOMAIN_CAN_POST`: Anyone in the account can post a message. " +
					"`ANYONE_CAN_POST`: Any Internet user who outside your account can access your Google Groups " +
					"service and post a message. " +
					"*Note: When `who_can_post_message` is set to `ANYONE_CAN_POST`, we recommend the" +
					"`message_moderation_level` be set to `MODERATE_NON_MEMBERS` to protect the group from possible spam. " +
					"Users not belonging to the organization are not allowed to become members of this group.",
				Type:     types.BoolType,
				Optional: true,
				Computed: true,
				Validators: []tfsdk.AttributeValidator{
					StringInSliceValidator{
						Options: []string{"NONE_CAN_POST",
							"ALL_MANAGERS_CAN_POST", "ALL_MEMBERS_CAN_POST", "ALL_OWNERS_CAN_POST", "ALL_IN_DOMAIN_CAN_POST",
							"ANYONE_CAN_POST"},
					},
				},
			},
			"allow_web_posting": {
				Description: "Allows posting from web. If true, allows any member to post to the group forum. If false, " +
					"Members only use Gmail to communicate with the group.",
				Type:     types.BoolType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.Bool{Value: true},
					},
				},
			},
			"primary_language": {
				Description: "The primary language for group. For a group's primary language use the language tags from " +
					"the Google Workspace languages found at Google Workspace Email Settings API Email Language Tags.",
				Type:     types.StringType,
				Optional: true,
			},
			"is_archived": {
				Description: "Allows the Group contents to be archived. If true, archive messages sent to the group. " +
					"If false, Do not keep an archive of messages sent to this group. If false, previously archived " +
					"messages remain in the archive.",
				Type:     types.BoolType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.Bool{Value: false},
					},
				},
			},
			"archive_only": {
				Description: "Allows the group to be archived only. If true, Group is archived and the group is inactive. " +
					"New messages to this group are rejected. The older archived messages are browsable and searchable. " +
					"If true, the `who_can_post_message` property is set to `NONE_CAN_POST`. If reverted from true to false, " +
					"`who_can_post_message` is set to `ALL_MANAGERS_CAN_POST`. If false, The group is active and can " +
					"receive messages. When false, updating `who_can_post_message` to `NONE_CAN_POST`, results in an error.",
				Type:     types.BoolType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.Bool{Value: false},
					},
				},
			},
			"message_moderation_level": {
				Description: "Moderation level of incoming messages. Possible values are: " +
					"`MODERATE_ALL_MESSAGES`: All messages are sent to the group owner's email address for approval. " +
					"If approved, the message is sent to the group. " +
					"`MODERATE_NON_MEMBERS`: All messages from non group members are sent to the group owner's email " +
					"address for approval. If approved, the message is sent to the group. " +
					"`MODERATE_NEW_MEMBERS`: All messages from new members are sent to the group owner's email address " +
					"for approval. If approved, the message is sent to the group. " +
					"`MODERATE_NONE`: No moderator approval is required. Messages are delivered directly to the group." +
					"Note: When the `who_can_post_message` is set to `ANYONE_CAN_POST`, we recommend the " +
					"`message_moderation_level` be set to `MODERATE_NON_MEMBERS` to protect the group from possible spam." +
					"When `member_can_post_as_the_group` is true, any message moderation settings on individual users " +
					"or new members will not apply to posts made on behalf of the group.",
				Type:     types.StringType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.String{Value: "MODERATE_NONE"},
					},
				},
				Validators: []tfsdk.AttributeValidator{
					StringInSliceValidator{
						Options: []string{"MODERATE_ALL_MESSAGES",
							"MODERATE_NON_MEMBERS", "MODERATE_NEW_MEMBERS", "MODERATE_NONE"},
					},
				},
			},
			"spam_moderation_level": {
				Description: "Specifies moderation levels for messages detected as spam. Possible values are: " +
					"`ALLOW`: Post the message to the group. " +
					"`MODERATE`: Send the message to the moderation queue. This is the default. " +
					"`SILENTLY_MODERATE`: Send the message to the moderation queue, but do not send notification to moderators. " +
					"`REJECT`: Immediately reject the message.",
				Type:     types.StringType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.String{Value: "MODERATE"},
					},
				},
				Validators: []tfsdk.AttributeValidator{
					StringInSliceValidator{
						Options: []string{"ALLOW",
							"MODERATE", "SILENTLY_MODERATE", "REJECT"},
					},
				},
			},
			"reply_to": {
				Description: "Specifies who receives the default reply. Possible values are: " +
					"`REPLY_TO_CUSTOM`: For replies to messages, use the group's custom email address. " +
					"When set to `REPLY_TO_CUSTOM`, the `custom_reply_to` property holds the custom email address used " +
					"when replying to a message, the customReplyTo property must have a value. Otherwise an error is returned. " +
					"`REPLY_TO_SENDER`: The reply sent to author of message. " +
					"`REPLY_TO_LIST`: This reply message is sent to the group. " +
					"`REPLY_TO_OWNER`: The reply is sent to the owner(s) of the group. This does not include the group's managers. " +
					"`REPLY_TO_IGNORE`: Group users individually decide where the message reply is sent. " +
					"`REPLY_TO_MANAGERS`: This reply message is sent to the group's managers, which includes all " +
					"managers and the group owner.",
				Type:     types.StringType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.String{Value: "REPLY_TO_IGNORE"},
					},
				},
				Validators: []tfsdk.AttributeValidator{
					StringInSliceValidator{
						Options: []string{"REPLY_TO_CUSTOM",
							"REPLY_TO_SENDER", "REPLY_TO_LIST", "REPLY_TO_OWNER", "REPLY_TO_IGNORE",
							"REPLY_TO_MANAGERS"},
					},
				},
			},
			"custom_reply_to": {
				Description: "An email address used when replying to a message if the `reply_to` property is set to " +
					"`REPLY_TO_CUSTOM`. This address is defined by an account administrator. When the group's `reply_to` " +
					"property is set to `REPLY_TO_CUSTOM`, the `custom_reply_to` property holds a custom email address " +
					"used when replying to a message, the `custom_reply_to` property must have a text value or an error is " +
					"returned.",
				Type:     types.StringType,
				Optional: true,
			},
			"include_custom_footer": {
				Description: "Whether to include custom footer.",
				Type:        types.BoolType,
				Optional:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.Bool{Value: false},
					},
				},
			},
			"custom_footer_text": {
				Description: "Set the content of custom footer text. The maximum number of characters is 1,000.",
				Type:        types.StringType,
				Optional:    true,
				Validators: []tfsdk.AttributeValidator{
					StringLenBetweenValidator{
						Min: 0,
						Max: 1000,
					},
				},
			},
			"send_message_deny_notification": {
				Description: "Allows a member to be notified if the member's message to the group is denied by the " +
					"group owner. If true, when a message is rejected, send the deny message notification to the " +
					"message author. The `default_message_deny_notification_text` property is dependent on the " +
					"`send_message_deny_notification` property being true. If false, when a message is rejected, " +
					"no notification is sent.",
				Type:     types.BoolType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.Bool{Value: false},
					},
				},
			},
			"default_message_deny_notification_text": {
				Description: "When a message is rejected, this is text for the rejection notification sent to the " +
					"message's author. By default, this property is empty and has no value in the API's response body. " +
					"The maximum notification text size is 10,000 characters. Requires `send_message_deny_notification` " +
					"property to be true.",
				Type:     types.StringType,
				Optional: true,
				Validators: []tfsdk.AttributeValidator{
					StringLenBetweenValidator{
						Min: 0,
						Max: 10000,
					},
				},
			},
			"members_can_post_as_the_group": {
				Description: "Enables members to post messages as the group. If true, group member can post messages " +
					"using the group's email address instead of their own email address. Message appear to originate " +
					"from the group itself. Any message moderation settings on individual users or new members do not " +
					"apply to posts made on behalf of the group. If false, members can not post in behalf of the " +
					"group's email address.",
				Type:     types.BoolType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.Bool{Value: false},
					},
				},
			},
			"include_in_global_address_list": {
				Description: "Enables the group to be included in the Global Address List. If true, the group is " +
					"included in the Global Address List. If false, it is not included in the Global Address List.",
				Type:     types.BoolType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.Bool{Value: true},
					},
				},
			},
			"who_can_leave_group": {
				Description: "Permission to leave the group. Possible values are: `ALL_MANAGERS_CAN_LEAVE`, " +
					"`ALL_MEMBERS_CAN_LEAVE`, `NONE_CAN_LEAVE`",
				Type:     types.StringType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.String{Value: "ALL_MEMBERS_CAN_LEAVE"},
					},
				},
				Validators: []tfsdk.AttributeValidator{
					StringInSliceValidator{
						Options: []string{"ALL_MANAGERS_CAN_LEAVE",
							"ALL_MEMBERS_CAN_LEAVE", "NONE_CAN_LEAVE"},
					},
				},
			},
			"who_can_contact_owner": {
				Description: "Permission to contact owner of the group via web UI. Possible values are: " +
					"`ALL_IN_DOMAIN_CAN_CONTACT`, `ALL_MANAGERS_CAN_CONTACT`, `ALL_MEMBERS_CAN_CONTACT`, " +
					"`ANYONE_CAN_CONTACT`, `ALL_OWNERS_CAN_CONTACT`",
				Type:     types.StringType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.String{Value: "ANYONE_CAN_CONTACT"},
					},
				},
				Validators: []tfsdk.AttributeValidator{
					StringInSliceValidator{
						Options: []string{"ALL_IN_DOMAIN_CAN_CONTACT",
							"ALL_MANAGERS_CAN_CONTACT", "ALL_MEMBERS_CAN_CONTACT", "ANYONE_CAN_CONTACT", "ALL_OWNERS_CAN_CONTACT"},
					},
				},
			},
			"who_can_moderate_members": {
				Description: "Specifies who can manage members. Possible values are: " +
					"`ALL_MEMBERS`, `OWNERS_AND_MANAGERS`, `OWNERS_ONLY`, `NONE`",
				Type:     types.StringType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.String{Value: "OWNERS_AND_MANAGERS"},
					},
				},
				Validators: []tfsdk.AttributeValidator{
					StringInSliceValidator{
						Options: []string{"ALL_MEMBERS",
							"OWNERS_AND_MANAGERS", "OWNERS_ONLY", "NONE"},
					},
				},
			},
			"who_can_moderate_content": {
				Description: "Specifies who can moderate content. Possible values are: " +
					"`ALL_MEMBERS`, `OWNERS_AND_MANAGERS`, `OWNERS_ONLY`, `NONE`",
				Type:     types.StringType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.String{Value: "OWNERS_AND_MANAGERS"},
					},
				},
				Validators: []tfsdk.AttributeValidator{
					StringInSliceValidator{
						Options: []string{"ALL_MEMBERS",
							"OWNERS_AND_MANAGERS", "OWNERS_ONLY", "NONE"},
					},
				},
			},
			"who_can_assist_content": {
				Description: "Specifies who can moderate metadata. Possible values are: " +
					"`ALL_MEMBERS`, `OWNERS_AND_MANAGERS`, `MANAGERS_ONLY`, `OWNERS_ONLY`, `NONE`",
				Type:     types.StringType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.String{Value: "NONE"},
					},
				},
				Validators: []tfsdk.AttributeValidator{
					StringInSliceValidator{
						Options: []string{"ALL_MEMBERS",
							"OWNERS_AND_MANAGERS", "MANAGERS_ONLY", "OWNERS_ONLY", "NONE"},
					},
				},
			},
			"custom_roles_enabled_for_settings_to_be_merged": {
				Description: "Specifies whether the group has a custom role that's included in one of the settings " +
					"being merged.",
				Type:     types.BoolType,
				Computed: true,
			},
			"enable_collaborative_inbox": {
				Description: "Specifies whether a collaborative inbox will remain turned on for the group.",
				Type:        types.BoolType,
				Optional:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.Bool{Value: false},
					},
				},
			},
			"who_can_discover_group": {
				Description: "Specifies the set of users for whom this group is discoverable. Possible values are: " +
					"`ANYONE_CAN_DISCOVER`, `ALL_IN_DOMAIN_CAN_DISCOVER`, `ALL_MEMBERS_CAN_DISCOVER`",
				Type:     types.StringType,
				Optional: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.String{Value: "ALL_IN_DOMAIN_CAN_DISCOVER"},
					},
				},
				Validators: []tfsdk.AttributeValidator{
					StringInSliceValidator{
						Options: []string{"ANYONE_CAN_DISCOVER",
							"ALL_IN_DOMAIN_CAN_DISCOVER", "ALL_MEMBERS_CAN_DISCOVER"},
					},
				},
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "Group Settings identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

type groupSettingsResource struct {
	provider provider
}

func (r resourceGroupSettingsType) NewResource(_ context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)

	return groupSettingsResource{
		provider: p,
	}, diags
}

// Create a new group settings
func (r groupSettingsResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from "+
				"another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan model.GroupSettingsResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupSettings := upsertGroupSettings(ctx, &r.provider, &plan, &resp.Diagnostics)

	diags = resp.State.Set(ctx, groupSettings)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Finished creating Group Settings %s", groupSettings.ID.Value)
}

// Read group settings information
func (r groupSettingsResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state model.GroupSettingsResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupSettings := GetGroupSettingsData(&r.provider, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if groupSettings.ID.Null {
		resp.State.RemoveResource(ctx)
		log.Printf("[DEBUG] Removed Org Unit from state because it was not found %s", state.ID.Value)
		return
	}

	diags = resp.State.Set(ctx, groupSettings)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	log.Printf("[DEBUG] Finished getting Group Settings %s", groupSettings.ID.Value)
}

// Update group settings
func (r groupSettingsResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Retrieve values from plan
	var plan model.GroupSettingsResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupSettings := upsertGroupSettings(ctx, &r.provider, &plan, &resp.Diagnostics)

	diags = resp.State.Set(ctx, groupSettings)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Finished updating Group Settings %s", groupSettings.ID.Value)
}

// Delete group settings
func (r groupSettingsResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state model.GroupSettingsResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Removing Group Settings %s from state", state.ID.Value)

	resp.State.RemoveResource(ctx)
	log.Printf("[DEBUG] Finished removing Group Settings from State: %s", state.ID.Value)
}

// ImportState group settings
func (r groupSettingsResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

func upsertGroupSettings(ctx context.Context, prov *provider, plan *model.GroupSettingsResourceData, diags *diag.Diagnostics) *model.GroupSettingsResourceData {
	groupSettingsReq := GroupSettingsPlanToObj(plan)

	log.Printf("[DEBUG] Creating Group Settings for %s", plan.Email.Value)
	groupsSettingsService := GetGroupsSettingsService(prov, diags)
	if diags.HasError() {
		return &model.GroupSettingsResourceData{}
	}

	groupSettingsObj, err := groupsSettingsService.Update(plan.Email.Value, &groupSettingsReq).Do()
	if err != nil {
		diags.AddError("error while trying to update group settings", err.Error())
		return &model.GroupSettingsResourceData{}
	}

	if groupSettingsObj == nil {
		diags.AddError(fmt.Sprintf("no org unit was returned for %s", plan.Name.Value), "object returned was nil")
		return &model.GroupSettingsResourceData{}
	}

	return SetGroupSettingsData(groupSettingsObj, diags)
}
