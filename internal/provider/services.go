package googleworkspace

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	directory "google.golang.org/api/admin/directory/v1"
)

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
