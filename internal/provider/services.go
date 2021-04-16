package googleworkspace

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	directory "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/groupssettings/v1"
)

func GetDomainsService(directoryService *directory.Service) (*directory.DomainsService, diag.Diagnostics) {
	var diags diag.Diagnostics

	log.Printf("[INFO] Instantiating Google Admin Domains service")
	domainsService := directoryService.Domains
	if domainsService == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Domains Service could not be created.",
		})

		return nil, diags
	}

	return domainsService, diags
}

func GetGroupsService(directoryService *directory.Service) (*directory.GroupsService, diag.Diagnostics) {
	var diags diag.Diagnostics

	log.Printf("[INFO] Instantiating Google Admin Groups service")
	groupsService := directoryService.Groups
	if groupsService == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Groups Service could not be created.",
		})

		return nil, diags
	}

	return groupsService, diags
}

func GetGroupsSettingsService(groupsSettingsService *groupssettings.Service) (*groupssettings.GroupsService, diag.Diagnostics) {
	var diags diag.Diagnostics

	log.Printf("[INFO] Instantiating Google Admin Groups Settings Groups service")
	groupsService := groupsSettingsService.Groups
	if groupsService == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Groups Service could not be created.",
		})

		return nil, diags
	}

	return groupsService, diags
}

func GetGroupAliasService(groupsService *directory.GroupsService) (*directory.GroupsAliasesService, diag.Diagnostics) {
	var diags diag.Diagnostics

	log.Printf("[INFO] Instantiating Google Admin Group Alias service")
	aliasesService := groupsService.Aliases
	if aliasesService == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Groups Aliases Service could not be created.",
		})

		return nil, diags
	}

	return aliasesService, diags
}

func GetMembersService(directoryService *directory.Service) (*directory.MembersService, diag.Diagnostics) {
	var diags diag.Diagnostics

	log.Printf("[INFO] Instantiating Google Admin Members service")
	membersService := directoryService.Members
	if membersService == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Members Service could not be created.",
		})

		return nil, diags
	}

	return membersService, diags
}

func GetSchemasService(directoryService *directory.Service) (*directory.SchemasService, diag.Diagnostics) {
	var diags diag.Diagnostics

	log.Printf("[INFO] Instantiating Google Admin Schemas service")
	schemasService := directoryService.Schemas
	if schemasService == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Schemas Service could not be created.",
		})

		return nil, diags
	}

	return schemasService, diags
}

func GetUsersService(directoryService *directory.Service) (*directory.UsersService, diag.Diagnostics) {
	var diags diag.Diagnostics

	log.Printf("[INFO] Instantiating Google Admin Users service")
	usersService := directoryService.Users
	if usersService == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Users Service could not be created.",
		})

		return nil, diags
	}

	return usersService, diags
}

func GetUserAliasService(usersService *directory.UsersService) (*directory.UsersAliasesService, diag.Diagnostics) {
	var diags diag.Diagnostics

	log.Printf("[INFO] Instantiating Google Admin User Alias service")
	aliasesService := usersService.Aliases
	if aliasesService == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Users Aliases Service could not be created.",
		})

		return nil, diags
	}

	return aliasesService, diags
}
