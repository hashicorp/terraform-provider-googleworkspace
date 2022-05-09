package model

import "github.com/hashicorp/terraform-plugin-framework/types"

type GroupSettingsResourceData struct {
	ID                                      types.String `tfsdk:"id"`
	Email                                   types.String `tfsdk:"email"`
	Name                                    types.String `tfsdk:"name"`
	Description                             types.String `tfsdk:"description"`
	WhoCanJoin                              types.String `tfsdk:"who_can_join"`
	WhoCanViewMembership                    types.String `tfsdk:"who_can_view_membership"`
	WhoCanViewGroup                         types.String `tfsdk:"who_can_view_group"`
	AllowExternalMembers                    types.Bool   `tfsdk:"allow_external_members"`
	WhoCanPostMessage                       types.String `tfsdk:"who_can_post_message"`
	AllowWebPosting                         types.Bool   `tfsdk:"allow_web_posting"`
	PrimaryLanguage                         types.String `tfsdk:"primary_language"`
	IsArchived                              types.Bool   `tfsdk:"is_archived"`
	ArchiveOnly                             types.Bool   `tfsdk:"archive_only"`
	MessageModerationLevel                  types.String `tfsdk:"message_moderation_level"`
	SpamModerationLevel                     types.String `tfsdk:"spam_moderation_level"`
	ReplyTo                                 types.String `tfsdk:"reply_to"`
	CustomReplyTo                           types.String `tfsdk:"custom_reply_to"`
	IncludeCustomFooter                     types.Bool   `tfsdk:"include_custom_footer"`
	CustomFooterText                        types.String `tfsdk:"custom_footer_text"`
	SendMessageDenyNotification             types.Bool   `tfsdk:"send_message_deny_notification"`
	DefaultMessageDenyNotificationText      types.String `tfsdk:"default_message_deny_notification_text"`
	MembersCanPostAsTheGroup                types.Bool   `tfsdk:"members_can_post_as__the_group"`
	IncludeInGlobalAddressList              types.Bool   `tfsdk:"include_in_global_address_list"`
	WhoCanLeaveGroup                        types.String `tfsdk:"who_can_leave_group"`
	WhoCanContactOwner                      types.String `tfsdk:"who_can_contact_owner"`
	WhoCanModerateMembers                   types.String `tfsdk:"who_can_moderate_members"`
	WhoCanModerateContent                   types.String `tfsdk:"who_can_moderate_content"`
	WhoCanAssistContent                     types.String `tfsdk:"who_can_assist_content"`
	CustomRolesEnabledForSettingsToBeMerged types.Bool   `tfsdk:"custom_roles_enabled_for_settings_to_be_merged"`
	EnableCollaborativeInbox                types.Bool   `tfsdk:"enable_collaborative_inbox"`
	WhoCanDiscoverGroup                     types.String `tfsdk:"who_can_discover_group"`
}
