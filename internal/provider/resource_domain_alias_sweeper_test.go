// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// This will sweep DomainAlias resources
func init() {
	resource.AddTestSweepers("DomainAlias", &resource.Sweeper{
		Name: "DomainAlias",
		F:    testSweepDomainAlias,
	})
}

// At the time of writing, the CI only passes us-central1 as the region
func testSweepDomainAlias(region string) error {
	resourceName := "DomainAlias"
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

	domainAliasesService, diags := GetDomainAliasesService(directoryService)
	if diags.HasError() {
		log.Printf("[INFO][SWEEPER_LOG] Error getting domain aliases service: %s", diags[0].Summary)
		return fmt.Errorf(diags[0].Summary)
	}

	domainAliases, err := domainAliasesService.List(client.Customer).Do()
	if err != nil {
		log.Printf("[INFO][SWEEPER_LOG] Error getting domain alias list: %s", err)
		return err
	}

	numDomainAliases := len(domainAliases.DomainAliases)
	if numDomainAliases == 0 {
		log.Printf("[INFO][SWEEPER_LOG] Nothing found in response.")
		return nil
	}

	log.Printf("[INFO][SWEEPER_LOG] Found %d items in %s list response.", numDomainAliases, resourceName)
	// Count items that weren't sweeped.
	nonPrefixCount := 0
	for _, domainAlias := range domainAliases.DomainAliases {
		// Increment count and skip if resource is not sweepable.
		if !isSweepableTestResource(domainAlias.DomainAliasName) {
			nonPrefixCount++
			continue
		}

		// Don't wait on operations as we may have a lot to delete
		err := domainAliasesService.Delete(client.Customer, domainAlias.DomainAliasName).Do()
		if err != nil {
			log.Printf("[INFO][SWEEPER_LOG] Error deleting domain alias: %s", domainAlias.DomainAliasName)
		} else {
			log.Printf("[INFO][SWEEPER_LOG] Sent delete request for %s resource: %s", resourceName, domainAlias.DomainAliasName)
		}
	}

	if nonPrefixCount > 0 {
		log.Printf("[INFO][SWEEPER_LOG] %d items without tf-test prefix remain.", nonPrefixCount)
	}

	return nil
}
