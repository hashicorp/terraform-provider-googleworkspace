// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// This will sweep OrgUnit resources
func init() {
	resource.AddTestSweepers("OrgUnit", &resource.Sweeper{
		Name: "OrgUnit",
		F:    testSweepOrgUnit,
	})
}

// At the time of writing, the CI only passes us-central1 as the region
func testSweepOrgUnit(region string) error {
	resourceName := "OrgUnit"
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

	orgUnitsService, diags := GetOrgUnitsService(directoryService)
	if diags.HasError() {
		log.Printf("[INFO][SWEEPER_LOG] Error getting orgUnits service: %s", diags[0].Summary)
		return fmt.Errorf(diags[0].Summary)
	}

	orgUnits, err := orgUnitsService.List(client.Customer).Do()
	if err != nil {
		log.Printf("[INFO][SWEEPER_LOG] Error getting orgUnit list: %s", err)
		return err
	}

	numOrgUnits := len(orgUnits.OrganizationUnits)
	if numOrgUnits == 0 {
		log.Printf("[INFO][SWEEPER_LOG] Nothing found in response.")
		return nil
	}

	log.Printf("[INFO][SWEEPER_LOG] Found %d items in %s list response.", numOrgUnits, resourceName)
	// Count items that weren't sweeped.
	nonPrefixCount := 0
	for _, orgUnit := range orgUnits.OrganizationUnits {
		// Increment count and skip if resource is not sweepable.
		if !isSweepableTestResource(orgUnit.Name) {
			nonPrefixCount++
			continue
		}

		// Don't wait on operations as we may have a lot to delete
		err := orgUnitsService.Delete(client.Customer, orgUnit.OrgUnitId).Do()
		if err != nil {
			log.Printf("[INFO][SWEEPER_LOG] Error deleting orgUnit: %s", orgUnit.Name)
		} else {
			log.Printf("[INFO][SWEEPER_LOG] Sent delete request for %s resource: %s", resourceName, orgUnit.Name)
		}
	}

	if nonPrefixCount > 0 {
		log.Printf("[INFO][SWEEPER_LOG] %d items without tf-test prefix remain.", nonPrefixCount)
	}

	return nil
}
