// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// This will sweep Role resources
func init() {
	resource.AddTestSweepers("Role", &resource.Sweeper{
		Name: "Role",
		F:    testSweepRole,
	})
}

// At the time of writing, the CI only passes us-central1 as the region
func testSweepRole(region string) error {
	resourceName := "Role"
	log.Printf("[INFO][SWEEPER_LOG] Starting sweeper for %s", resourceName)

	client, err := googleworkspaceTestClient()
	if err != nil {
		log.Printf("[INFO][SWEEPER_LOG] error getting sweeper client: %s", err)
		return err
	}

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		log.Printf("[INFO][SWEEPER_LOG] Error creating directory service: %s", diags[0].Summary)
		return fmt.Errorf(diags[0].Summary)
	}

	rolesService, diags := GetRolesService(directoryService)
	if diags.HasError() {
		log.Printf("[INFO][SWEEPER_LOG] Error getting Roles service: %s", diags[0].Summary)
		return fmt.Errorf(diags[0].Summary)
	}

	roles, err := rolesService.List(client.Customer).Do()
	if err != nil {
		log.Printf("[INFO][SWEEPER_LOG] Error getting Roles list: %s", err)
		return err
	}

	numRoles := len(roles.Items)
	if numRoles == 0 {
		log.Printf("[INFO][SWEEPER_LOG] Nothing found in response.")
		return nil
	}

	log.Printf("[INFO][SWEEPER_LOG] Found %d items in %s list response.", numRoles, resourceName)
	// Count items that weren't sweeped.
	nonPrefixCount := 0
	for _, role := range roles.Items {
		// Increment count and skip if resource is not sweepable.
		if !isSweepableTestResource(role.RoleName) {
			nonPrefixCount++
			continue
		}

		// Don't wait on operations as we may have a lot to delete
		err := rolesService.Delete(client.Customer, strconv.FormatInt(role.RoleId, 10)).Do()
		if err != nil {
			log.Printf("[INFO][SWEEPER_LOG] Error deleting Role: %s", role.RoleName)
		} else {
			log.Printf("[INFO][SWEEPER_LOG] Sent delete request for %s resource: %s", resourceName, role.RoleName)
		}
	}

	if nonPrefixCount > 0 {
		log.Printf("[INFO][SWEEPER_LOG] %d items without tf-test prefix remain.", nonPrefixCount)
	}

	return nil
}
