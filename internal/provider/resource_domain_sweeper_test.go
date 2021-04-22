package googleworkspace

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// This will sweep Domain resources
func init() {
	resource.AddTestSweepers("Domain", &resource.Sweeper{
		Name: "Domain",
		F:    testSweepDomain,
	})
}

// At the time of writing, the CI only passes us-central1 as the region
func testSweepDomain(region string) error {
	resourceName := "Domain"
	log.Printf("[INFO][SWEEPER_LOG] Starting sweeper for %s", resourceName)

	client, err := sharedClientForRegion()
	if err != nil {
		log.Printf("[INFO][SWEEPER_LOG] error getting sweeper client: %s", err)
		return err
	}

	diags := client.loadAndValidate(context.Background())
	if diags.HasError() {
		log.Printf("[INFO][SWEEPER_LOG] error loading: %s", diags[0].Summary)
		return fmt.Errorf(diags[0].Summary)
	}

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		log.Printf("[INFO][SWEEPER_LOG] Error creating directory service: %s", diags[0].Summary)
		return fmt.Errorf(diags[0].Summary)
	}

	domainsService, diags := GetDomainsService(directoryService)
	if diags.HasError() {
		log.Printf("[INFO][SWEEPER_LOG] Error getting domains service: %s", diags[0].Summary)
		return fmt.Errorf(diags[0].Summary)
	}

	domains, err := domainsService.List(client.Customer).Do()
	if err != nil {
		log.Printf("[INFO][SWEEPER_LOG] Error getting domain list: %s", err)
		return err
	}

	numDomains := len(domains.Domains)
	if numDomains == 0 {
		log.Printf("[INFO][SWEEPER_LOG] Nothing found in response.")
		return nil
	}

	log.Printf("[INFO][SWEEPER_LOG] Found %d items in %s list response.", numDomains, resourceName)
	// Count items that weren't sweeped.
	nonPrefixCount := 0
	for _, domain := range domains.Domains {
		// Increment count and skip if resource is not sweepable.
		if !isSweepableTestResource(domain.DomainName) {
			nonPrefixCount++
			continue
		}

		// Don't wait on operations as we may have a lot to delete
		err := domainsService.Delete(client.Customer, domain.DomainName).Do()
		if err != nil {
			log.Printf("[INFO][SWEEPER_LOG] Error deleting domain: %s", domain.DomainName)
		} else {
			log.Printf("[INFO][SWEEPER_LOG] Sent delete request for %s resource: %s", resourceName, domain.DomainName)
		}
	}

	if nonPrefixCount > 0 {
		log.Printf("[INFO][SWEEPER_LOG] %d items without tf-test prefix remain.", nonPrefixCount)
	}

	return nil
}
