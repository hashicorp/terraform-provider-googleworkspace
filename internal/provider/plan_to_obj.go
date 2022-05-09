package googleworkspace

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-googleworkspace-pf/internal/model"
	directory "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/groupssettings/v1"
	"strconv"
)

func UserPlanToObj(ctx context.Context, prov *provider, config, plan *model.UserResourceData, diags *diag.Diagnostics) directory.User {
	// All the lists
	emails := emailsToInterfaces(ctx, config.Emails.Elems, diags)
	externalIds := externalIdsToInterfaces(ctx, config.Emails.Elems, diags)
	relations := relationsToInterfaces(ctx, config.Emails.Elems, diags)
	addresses := addressesToInterfaces(ctx, config.Emails.Elems, diags)
	organizations := organizationsToInterfaces(ctx, config.Emails.Elems, diags)
	phones := phonesToInterfaces(ctx, config.Emails.Elems, diags)
	languages := languagesToInterfaces(ctx, config.Emails.Elems, diags)
	posixAccounts := posixAccountsToInterfaces(ctx, config.Emails.Elems, diags)
	sshPublicKeys := sshPublicKeysToInterfaces(ctx, config.Emails.Elems, diags)
	websites := websitesToInterfaces(ctx, config.Emails.Elems, diags)
	locations := locationsToInterfaces(ctx, config.Emails.Elems, diags)
	keywords := keywordsToInterfaces(ctx, config.Emails.Elems, diags)
	ims := imsToInterfaces(ctx, config.Emails.Elems, diags)
	//customSchemas :=

	var userName model.UserResourceName
	d := plan.Name.As(ctx, &userName, types.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return directory.User{}
	}

	return directory.User{
		PrimaryEmail:              plan.PrimaryEmail.Value,
		Password:                  config.Password.Value,
		HashFunction:              plan.HashFunction.Value,
		Suspended:                 plan.Suspended.Value,
		ChangePasswordAtNextLogin: plan.ChangePasswordAtNextLogin.Value,
		IpWhitelisted:             plan.IpAllowlist.Value,
		Name: &directory.UserName{
			FamilyName: userName.FamilyName.Value,
			GivenName:  userName.GivenName.Value,
		},
		Emails:                     emails,
		ExternalIds:                externalIds,
		Relations:                  relations,
		Aliases:                    typeListToSliceStrings(plan.Aliases.Elems),
		Addresses:                  addresses,
		Organizations:              organizations,
		Phones:                     phones,
		Languages:                  languages,
		PosixAccounts:              posixAccounts,
		SshPublicKeys:              sshPublicKeys,
		Websites:                   websites,
		Locations:                  locations,
		IncludeInGlobalAddressList: plan.IncludeInGlobalAddressList.Value,
		Keywords:                   keywords,
		Ims:                        ims,
		//CustomSchemas:              customSchemas,
		Archived:      plan.Archived.Value,
		OrgUnitPath:   plan.OrgUnitPath.Value,
		RecoveryEmail: plan.RecoveryEmail.Value,
		RecoveryPhone: plan.RecoveryPhone.Value,
	}
}

func DomainPlanToObj(plan *model.DomainResourceData) directory.Domains {
	return directory.Domains{
		DomainName: plan.DomainName.Value,
	}
}

func DomainAliasPlanToObj(plan *model.DomainAliasResourceData) directory.DomainAlias {
	return directory.DomainAlias{
		ParentDomainName: plan.ParentDomainName.Value,
		DomainAliasName:  plan.DomainAliasName.Value,
	}
}

func GmailSendAsAliasPlanToObj(ctx context.Context, plan *model.GmailSendAsAliasResourceData, diags *diag.Diagnostics) gmail.SendAs {
	var smtpMsa *model.GmailSendAsAliasResourceSmtpMsa
	d := plan.SmtpMsa.As(ctx, &smtpMsa, types.ObjectAsOptions{})
	diags.Append(d...)
	if diags.HasError() {
		return gmail.SendAs{}
	}

	return gmail.SendAs{
		SendAsEmail:    plan.SendAsEmail.Value,
		DisplayName:    plan.DisplayName.Value,
		ReplyToAddress: plan.ReplyToAddress.Value,
		Signature:      plan.Signature.Value,
		IsDefault:      plan.IsDefault.Value,
		TreatAsAlias:   plan.TreatAsAlias.Value,
		SmtpMsa: &gmail.SmtpMsa{
			Host:         smtpMsa.Host.Value,
			Port:         smtpMsa.Port.Value,
			Username:     smtpMsa.Username.Value,
			Password:     smtpMsa.Password.Value,
			SecurityMode: smtpMsa.SecurityMode.Value,
		},
	}
}

func GroupPlanToObj(plan *model.GroupResourceData) directory.Group {
	return directory.Group{
		Email:       plan.Email.Value,
		Name:        plan.Name.Value,
		Description: plan.Description.Value,
	}
}

func GroupMemberPlanToObj(plan *model.GroupMemberResourceData) directory.Member {
	return directory.Member{
		Email:            plan.Email.Value,
		Role:             plan.Role.Value,
		Type:             plan.Type.Value,
		DeliverySettings: plan.DeliverySettings.Value,
	}
}

func GroupMembersPlanToObj(ctx context.Context, plan *model.GroupMembersResourceData, diags *diag.Diagnostics) []directory.Member {
	members := []directory.Member{}
	for _, m := range plan.Members.Elems {
		var mem *model.GroupMembersResourceMember
		d := m.(types.Object).As(ctx, mem, types.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return []directory.Member{}
		}

		members = append(members, directory.Member{
			Email:            mem.Email.Value,
			Role:             mem.Role.Value,
			Type:             mem.Type.Value,
			DeliverySettings: mem.DeliverySettings.Value,
		})
	}

	return members
}

func GroupSettingsPlanToObj(plan *model.GroupSettingsResourceData) groupssettings.Groups {
	return groupssettings.Groups{
		Email:                                   plan.Email.Value,
		Name:                                    plan.Name.Value,
		Description:                             plan.Description.Value,
		WhoCanJoin:                              plan.WhoCanJoin.Value,
		WhoCanViewMembership:                    plan.WhoCanViewMembership.Value,
		WhoCanViewGroup:                         plan.WhoCanViewGroup.Value,
		AllowExternalMembers:                    strconv.FormatBool(plan.AllowExternalMembers.Value),
		WhoCanPostMessage:                       plan.WhoCanPostMessage.Value,
		AllowWebPosting:                         strconv.FormatBool(plan.AllowWebPosting.Value),
		PrimaryLanguage:                         plan.PrimaryLanguage.Value,
		IsArchived:                              strconv.FormatBool(plan.IsArchived.Value),
		ArchiveOnly:                             strconv.FormatBool(plan.ArchiveOnly.Value),
		MessageModerationLevel:                  plan.MessageModerationLevel.Value,
		SpamModerationLevel:                     plan.SpamModerationLevel.Value,
		ReplyTo:                                 plan.ReplyTo.Value,
		CustomReplyTo:                           plan.CustomReplyTo.Value,
		IncludeCustomFooter:                     strconv.FormatBool(plan.IncludeCustomFooter.Value),
		CustomFooterText:                        plan.CustomFooterText.Value,
		SendMessageDenyNotification:             strconv.FormatBool(plan.SendMessageDenyNotification.Value),
		DefaultMessageDenyNotificationText:      plan.DefaultMessageDenyNotificationText.Value,
		MembersCanPostAsTheGroup:                strconv.FormatBool(plan.MembersCanPostAsTheGroup.Value),
		IncludeInGlobalAddressList:              strconv.FormatBool(plan.IncludeInGlobalAddressList.Value),
		WhoCanLeaveGroup:                        plan.WhoCanLeaveGroup.Value,
		WhoCanContactOwner:                      plan.WhoCanContactOwner.Value,
		WhoCanModerateMembers:                   plan.WhoCanModerateMembers.Value,
		WhoCanModerateContent:                   plan.WhoCanModerateContent.Value,
		WhoCanAssistContent:                     plan.WhoCanAssistContent.Value,
		CustomRolesEnabledForSettingsToBeMerged: strconv.FormatBool(plan.CustomRolesEnabledForSettingsToBeMerged.Value),
		EnableCollaborativeInbox:                strconv.FormatBool(plan.EnableCollaborativeInbox.Value),
		WhoCanDiscoverGroup:                     plan.WhoCanDiscoverGroup.Value,

		ForceSendFields: []string{"AllowExternalMembers", "AllowWebPosting", "IsArchived", "ArchiveOnly",
			"IncludeCustomFooter", "SendMessageDenyNotification", "MembersCanPostAsTheGroup", "IncludeInGlobalAddressList",
			"CustomRolesEnabledForSettingsToBeMerged", "EnableCollaborativeInbox"},
	}
}

func OrgUnitPlanToObj(plan *model.OrgUnitResourceData) directory.OrgUnit {
	return directory.OrgUnit{
		Name:              plan.Name.Value,
		Description:       plan.Description.Value,
		BlockInheritance:  plan.BlockInheritance.Value,
		ParentOrgUnitId:   plan.ParentOrgUnitId.Value,
		ParentOrgUnitPath: plan.ParentOrgUnitPath.Value,
	}
}

func RolePlanToObj(ctx context.Context, plan *model.RoleResourceData, diags *diag.Diagnostics) directory.Role {
	rolePrivileges := []*directory.RoleRolePrivileges{}
	for _, p := range plan.Privileges.Elems {
		var priv model.RoleResourcePrivilege
		d := p.(types.Object).As(ctx, &priv, types.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return directory.Role{}
		}

		rolePrivileges = append(rolePrivileges, &directory.RoleRolePrivileges{
			PrivilegeName: priv.PrivilegeName.Value,
			ServiceId:     priv.ServiceId.Value,
		})
	}

	role := directory.Role{
		RoleName:        plan.Name.Value,
		RoleDescription: plan.Description.Value,
		RolePrivileges:  rolePrivileges,
	}

	if !plan.ID.Null {
		id, _ := strconv.ParseInt(plan.ID.Value, 10, 64)
		role.RoleId = id
	}
	return role
}

func RoleAssignmentPlanToObj(plan *model.RoleAssignmentResourceData) directory.RoleAssignment {
	return directory.RoleAssignment{
		RoleId:     plan.RoleId.Value,
		AssignedTo: plan.AssignedTo.Value,
		ScopeType:  plan.ScopeType.Value,
		OrgUnitId:  plan.OrgUnitId.Value,
	}
}

func SchemaPlanToObj(ctx context.Context, plan *model.SchemaResourceData, diags *diag.Diagnostics) directory.Schema {
	fields := []*directory.SchemaFieldSpec{}
	for _, f := range plan.Fields.Elems {
		var field model.SchemaResourceField
		d := f.(types.Object).As(ctx, &field, types.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return directory.Schema{}
		}

		var nis model.SchemaResourceFieldNumericIndexingSpec
		d = f.(types.Object).As(ctx, &nis, types.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return directory.Schema{}
		}

		fields = append(fields, &directory.SchemaFieldSpec{
			FieldName:      field.FieldName.Value,
			FieldId:        field.FieldId.Value,
			FieldType:      field.FieldType.Value,
			MultiValued:    field.MultiValued.Value,
			Indexed:        &field.Indexed.Value,
			DisplayName:    field.DisplayName.Value,
			ReadAccessType: field.ReadAccessType.Value,
			NumericIndexingSpec: &directory.SchemaFieldSpecNumericIndexingSpec{
				MinValue: nis.MinValue.Value,
				MaxValue: nis.MaxValue.Value,
			},
		})
	}

	return directory.Schema{
		SchemaName:  plan.SchemaName.Value,
		DisplayName: plan.DisplayName.Value,
		Fields:      fields,
	}
}
