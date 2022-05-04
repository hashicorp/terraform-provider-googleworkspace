package googleworkspace

import (
	"log"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	directory "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/chromepolicy/v1"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/groupssettings/v1"
)

func GetChromePoliciesService(prov *provider, diags *diag.Diagnostics) *chromepolicy.CustomersPoliciesService {
	chromePolicyService := prov.NewChromePolicyService(diags)

	log.Printf("[INFO] Instantiating Google Admin Chrome Policies service")
	customersService := chromePolicyService.Customers
	if customersService == nil || customersService.Policies == nil {
		diags.AddError("Chrome Policies Service could not be created.", "returned service was null.")
		return nil
	}

	return customersService.Policies
}

func GetChromePolicySchemasService(prov *provider, diags *diag.Diagnostics) *chromepolicy.CustomersPolicySchemasService {
	chromePolicyService := prov.NewChromePolicyService(diags)

	log.Printf("[INFO] Instantiating Google Admin Chrome Policy Schemas service")
	customersService := chromePolicyService.Customers
	if customersService == nil || customersService.PolicySchemas == nil {
		diags.AddError("Chrome Policies Schemas Service could not be created.", "returned service was null.")
		return nil
	}

	return customersService.PolicySchemas
}

func GetDomainAliasesService(prov *provider, diags *diag.Diagnostics) *directory.DomainAliasesService {
	directoryService := prov.NewDirectoryService(diags)

	log.Printf("[INFO] Instantiating Google Admin Domain Aliases service")
	domainAliasesService := directoryService.DomainAliases
	if domainAliasesService == nil {
		diags.AddError("Domain Aliases Service could not be created.", "returned service was null.")
		return nil
	}

	return domainAliasesService
}

func GetDomainsService(prov *provider, diags *diag.Diagnostics) *directory.DomainsService {
	directoryService := prov.NewDirectoryService(diags)

	log.Printf("[INFO] Instantiating Google Admin Domains service")
	domainsService := directoryService.Domains
	if domainsService == nil {
		diags.AddError("Domains Service could not be created.", "returned service was null.")
		return nil
	}

	return domainsService
}

func GetGroupsService(prov *provider, diags *diag.Diagnostics) *directory.GroupsService {
	directoryService := prov.NewDirectoryService(diags)

	log.Printf("[INFO] Instantiating Google Admin Groups service")
	groupsService := directoryService.Groups
	if groupsService == nil {
		diags.AddError("Groups Service could not be created.", "returned service was null.")
		return nil
	}

	return groupsService
}

func GetGroupsSettingsService(prov *provider, diags *diag.Diagnostics) *groupssettings.GroupsService {
	groupsSettingsService := prov.NewGroupsSettingsService(diags)

	log.Printf("[INFO] Instantiating Google Admin Groups Settings Groups service")
	groupsService := groupsSettingsService.Groups
	if groupsService == nil {
		diags.AddError("Groups Settings Groups Service could not be created.", "returned service was null.")
		return nil
	}

	return groupsService
}

func GetGmailSendAsAliasService(prov *provider, diags *diag.Diagnostics) *gmail.UsersSettingsSendAsService {
	gmailService := prov.NewGmailService(diags)

	log.Printf("[INFO] Instantiating Google Admin Gmail Send As Alias service")
	usersService := gmailService.Users
	if usersService == nil || usersService.Settings == nil || usersService.Settings.SendAs == nil {
		diags.AddError("Gmail Send As Alias Service could not be created.", "returned service was null.")
		return nil
	}

	return usersService.Settings.SendAs
}

func GetGroupAliasService(prov *provider, diags *diag.Diagnostics) *directory.GroupsAliasesService {
	groupsService := GetGroupsService(prov, diags)

	log.Printf("[INFO] Instantiating Google Admin Group Alias service")
	aliasesService := groupsService.Aliases
	if aliasesService == nil {
		diags.AddError("Group Alias Service could not be created.", "returned service was null.")
		return nil
	}

	return aliasesService
}

func GetMembersService(prov *provider, diags *diag.Diagnostics) *directory.MembersService {
	directoryService := prov.NewDirectoryService(diags)

	log.Printf("[INFO] Instantiating Google Admin Members service")
	membersService := directoryService.Members
	if membersService == nil {
		diags.AddError("Members Service could not be created.", "returned service was null.")
		return nil
	}

	return membersService
}

func GetOrgUnitsService(prov *provider, diags *diag.Diagnostics) *directory.OrgunitsService {
	directoryService := prov.NewDirectoryService(diags)

	log.Printf("[INFO] Instantiating Google Admin OrgUnits service")
	ousService := directoryService.Orgunits
	if ousService == nil {
		diags.AddError("OrgUnits Service could not be created.", "returned service was null.")
		return nil
	}

	return ousService
}

func GetPrivilegesService(prov *provider, diags *diag.Diagnostics) *directory.PrivilegesService {
	directoryService := prov.NewDirectoryService(diags)

	log.Printf("[INFO] Instantiating Google Admin Privileges service")
	privilegesService := directoryService.Privileges
	if privilegesService == nil {
		diags.AddError("Privileges Service could not be created.", "returned service was null.")
		return nil
	}

	return privilegesService
}

func GetRoleAssignmentsService(prov *provider, diags *diag.Diagnostics) *directory.RoleAssignmentsService {
	directoryService := prov.NewDirectoryService(diags)

	log.Printf("[INFO] Instantiating Google Admin RoleAssignments service")
	roleAssignmentsService := directoryService.RoleAssignments
	if roleAssignmentsService == nil {
		diags.AddError("RoleAssignments Service could not be created.", "returned service was null.")
		return nil
	}

	return roleAssignmentsService
}

func GetRolesService(prov *provider, diags *diag.Diagnostics) *directory.RolesService {
	directoryService := prov.NewDirectoryService(diags)

	log.Printf("[INFO] Instantiating Google Admin Roles service")
	rolesService := directoryService.Roles
	if rolesService == nil {
		diags.AddError("Roles Service could not be created.", "returned service was null.")
		return nil
	}

	return rolesService
}

func GetSchemasService(prov *provider, diags *diag.Diagnostics) *directory.SchemasService {
	directoryService := prov.NewDirectoryService(diags)

	log.Printf("[INFO] Instantiating Google Admin Schemas service")
	schemasService := directoryService.Schemas
	if schemasService == nil {
		diags.AddError("Schemas Service could not be created.", "returned service was null.")
		return nil
	}

	return schemasService
}

func GetUsersService(prov *provider, diags *diag.Diagnostics) *directory.UsersService {
	directoryService := prov.NewDirectoryService(diags)

	log.Printf("[INFO] Instantiating Google Admin Users service")
	usersService := directoryService.Users
	if usersService == nil {
		diags.AddError("Users Service could not be created.", "returned service was null.")
		return nil
	}

	return usersService
}

func GetUserAliasService(prov *provider, diags *diag.Diagnostics) *directory.UsersAliasesService {
	usersService := GetUsersService(prov, diags)

	log.Printf("[INFO] Instantiating Google Admin User Aliases service")
	aliasesService := usersService.Aliases
	if aliasesService == nil {
		diags.AddError("User Aliases Service could not be created.", "returned service was null.")
		return nil
	}

	return aliasesService
}
