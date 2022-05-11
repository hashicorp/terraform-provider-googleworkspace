package googleworkspace

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-provider-googleworkspace-pf/internal/model"
	"google.golang.org/api/groupssettings/v1"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type datasourceGroupSettingsType struct{}

func (t datasourceGroupSettingsType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	resourceType := resourceDomainType{}

	resourceSchema, diags := resourceType.GetSchema(ctx)
	if diags.HasError() {
		return tfsdk.Schema{}, diags
	}

	attrs := datasourceSchemaFromResourceSchema(resourceSchema.Attributes)
	addRequiredFieldsToSchema(attrs, "email")

	return tfsdk.Schema{
		Description: "Group Settings data source in the Terraform Googleworkspace provider. Group Settings resides " +
			"under the `https://www.googleapis.com/auth/apps.groups.settings` client scope.",
		Attributes: attrs,
	}, nil
}

type groupSettingsDatasource struct {
	provider provider
}

func (t datasourceGroupSettingsType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	p, diags := convertProviderType(in)

	return groupSettingsDatasource{
		provider: p,
	}, diags
}

func (d groupSettingsDatasource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data model.GroupSettingsResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupSettings := GetGroupSettingsData(&d.provider, &data, &resp.Diagnostics)
	if groupSettings.ID.Value == "" {
		resp.Diagnostics.AddError("object does not exist",
			fmt.Sprintf("Group Settings %s does not exist", data.Email.Value))
	}

	diags = resp.State.Set(ctx, groupSettings)
	resp.Diagnostics.Append(diags...)
}

func GetGroupSettingsData(prov *provider, plan *model.GroupSettingsResourceData, diags *diag.Diagnostics) *model.GroupSettingsResourceData {
	groupsSettingsService := GetGroupsSettingsService(prov, diags)
	log.Printf("[DEBUG] Getting Group Settings %s", plan.Email.Value)

	groupSettingsObj, err := groupsSettingsService.Get(plan.Email.Value).Do()
	if err != nil {
		plan.ID.Value = handleNotFoundError(err, plan.ID.Value, diags)
	}

	if groupSettingsObj == nil {
		diags.AddError("returned obj is nil", fmt.Sprintf("GET %s returned nil object",
			plan.Email.Value))
	}

	return SetGroupSettingsData(groupSettingsObj, diags)
}

func SetGroupSettingsData(obj *groupssettings.Groups, diags *diag.Diagnostics) *model.GroupSettingsResourceData {
	// Parse Bools
	allowExternalMembers, err := strconv.ParseBool(obj.AllowExternalMembers)
	if err != nil {
		diags.AddError("error converting string to bool", err.Error())
	}
	allowWebPosting, err := strconv.ParseBool(obj.AllowWebPosting)
	if err != nil {
		diags.AddError("error converting string to bool", err.Error())
	}
	isArchived, err := strconv.ParseBool(obj.IsArchived)
	if err != nil {
		diags.AddError("error converting string to bool", err.Error())
	}
	archiveOnly, err := strconv.ParseBool(obj.ArchiveOnly)
	if err != nil {
		diags.AddError("error converting string to bool", err.Error())
	}
	includeCustomFooter, err := strconv.ParseBool(obj.IncludeCustomFooter)
	if err != nil {
		diags.AddError("error converting string to bool", err.Error())
	}
	sendMessageDenyNotification, err := strconv.ParseBool(obj.SendMessageDenyNotification)
	if err != nil {
		diags.AddError("error converting string to bool", err.Error())
	}
	membersCanPostAsTheGroup, err := strconv.ParseBool(obj.MembersCanPostAsTheGroup)
	if err != nil {
		diags.AddError("error converting string to bool", err.Error())
	}
	includeInGlobalAddressList, err := strconv.ParseBool(obj.IncludeInGlobalAddressList)
	if err != nil {
		diags.AddError("error converting string to bool", err.Error())
	}
	customRolesEnabledForSettingsToBeMerged, err := strconv.ParseBool(obj.CustomRolesEnabledForSettingsToBeMerged)
	if err != nil {
		diags.AddError("error converting string to bool", err.Error())
	}
	enableCollaborativeInbox, err := strconv.ParseBool(obj.EnableCollaborativeInbox)
	if err != nil {
		diags.AddError("error converting string to bool", err.Error())
	}
	if diags.HasError() {
		return &model.GroupSettingsResourceData{}
	}

	return &model.GroupSettingsResourceData{
		ID:                                      types.String{Value: obj.Email},
		Email:                                   types.String{Value: obj.Email},
		Name:                                    types.String{Value: obj.Name},
		Description:                             types.String{Value: obj.Description},
		WhoCanJoin:                              types.String{Value: obj.WhoCanJoin},
		WhoCanViewMembership:                    types.String{Value: obj.WhoCanViewMembership},
		WhoCanViewGroup:                         types.String{Value: obj.WhoCanViewGroup},
		AllowExternalMembers:                    types.Bool{Value: allowExternalMembers},
		WhoCanPostMessage:                       types.String{Value: obj.WhoCanPostMessage},
		AllowWebPosting:                         types.Bool{Value: allowWebPosting},
		PrimaryLanguage:                         types.String{Value: obj.PrimaryLanguage},
		IsArchived:                              types.Bool{Value: isArchived},
		ArchiveOnly:                             types.Bool{Value: archiveOnly},
		MessageModerationLevel:                  types.String{Value: obj.MessageModerationLevel},
		SpamModerationLevel:                     types.String{Value: obj.SpamModerationLevel},
		ReplyTo:                                 types.String{Value: obj.ReplyTo},
		CustomReplyTo:                           types.String{Value: obj.ReplyTo},
		IncludeCustomFooter:                     types.Bool{Value: includeCustomFooter},
		CustomFooterText:                        types.String{Value: obj.CustomFooterText},
		SendMessageDenyNotification:             types.Bool{Value: sendMessageDenyNotification},
		DefaultMessageDenyNotificationText:      types.String{Value: obj.DefaultMessageDenyNotificationText},
		MembersCanPostAsTheGroup:                types.Bool{Value: membersCanPostAsTheGroup},
		IncludeInGlobalAddressList:              types.Bool{Value: includeInGlobalAddressList},
		WhoCanLeaveGroup:                        types.String{Value: obj.WhoCanLeaveGroup},
		WhoCanContactOwner:                      types.String{Value: obj.WhoCanContactOwner},
		WhoCanModerateMembers:                   types.String{Value: obj.WhoCanModerateMembers},
		WhoCanModerateContent:                   types.String{Value: obj.WhoCanModerateContent},
		WhoCanAssistContent:                     types.String{Value: obj.WhoCanAssistContent},
		CustomRolesEnabledForSettingsToBeMerged: types.Bool{Value: customRolesEnabledForSettingsToBeMerged},
		EnableCollaborativeInbox:                types.Bool{Value: enableCollaborativeInbox},
		WhoCanDiscoverGroup:                     types.String{Value: obj.WhoCanDiscoverGroup},
	}
}

//func dataSourceGroupSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//
//	d.SetId(d.Get("email").(string))
//
//	return resourceGroupSettingsRead(ctx, d, meta)
//}
