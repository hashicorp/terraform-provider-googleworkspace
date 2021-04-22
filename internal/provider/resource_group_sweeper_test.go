package googleworkspace

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// This will sweep Group resources
func init() {
	resource.AddTestSweepers("Group", &resource.Sweeper{
		Name: "Group",
		F:    testSweepGroup,
	})
}

// At the time of writing, the CI only passes us-central1 as the region
func testSweepGroup(region string) error {
	resourceName := "Group"
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

	groupsService, diags := GetGroupsService(directoryService)
	if diags.HasError() {
		log.Printf("[INFO][SWEEPER_LOG] Error getting groups service: %s", diags[0].Summary)
		return fmt.Errorf(diags[0].Summary)
	}

	groups, err := groupsService.List().Customer(client.Customer).Do()
	if err != nil {
		log.Printf("[INFO][SWEEPER_LOG] Error getting group list: %s", err)
		return err
	}

	numGroups := len(groups.Groups)
	if numGroups == 0 {
		log.Printf("[INFO][SWEEPER_LOG] Nothing found in response.")
		return nil
	}

	log.Printf("[INFO][SWEEPER_LOG] Found %d items in %s list response.", numGroups, resourceName)
	// Count items that weren't sweeped.
	nonPrefixCount := 0
	for _, group := range groups.Groups {
		// Increment count and skip if resource is not sweepable.
		if !isSweepableTestResource(group.Email) {
			nonPrefixCount++
			continue
		}

		// Don't wait on operations as we may have a lot to delete
		err := groupsService.Delete(group.Id).Do()
		if err != nil {
			log.Printf("[INFO][SWEEPER_LOG] Error deleting group: %s", group.Email)
		} else {
			log.Printf("[INFO][SWEEPER_LOG] Sent delete request for %s resource: %s", resourceName, group.Email)
		}
	}

	if nonPrefixCount > 0 {
		log.Printf("[INFO][SWEEPER_LOG] %d items without tf-test prefix remain.", nonPrefixCount)
	}

	return nil
}
