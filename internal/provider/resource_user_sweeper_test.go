package googleworkspace

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// This will sweep User resources
func init() {
	resource.AddTestSweepers("User", &resource.Sweeper{
		Name: "User",
		F:    testSweepUser,
	})
}

// At the time of writing, the CI only passes us-central1 as the region
func testSweepUser(region string) error {
	resourceName := "User"
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

	usersService, diags := GetUsersService(directoryService)
	if diags.HasError() {
		log.Printf("[INFO][SWEEPER_LOG] Error getting users service: %s", diags[0].Summary)
		return fmt.Errorf(diags[0].Summary)
	}

	users, err := usersService.List().Customer(client.Customer).Do()
	if err != nil {
		log.Printf("[INFO][SWEEPER_LOG] Error getting user list: %s", err)
		return err
	}

	numUsers := len(users.Users)
	if numUsers == 0 {
		log.Printf("[INFO][SWEEPER_LOG] Nothing found in response.")
		return nil
	}

	log.Printf("[INFO][SWEEPER_LOG] Found %d items in %s list response.", numUsers, resourceName)
	// Count items that weren't sweeped.
	nonPrefixCount := 0
	for _, user := range users.Users {
		// Increment count and skip if resource is not sweepable.
		if !isSweepableTestResource(user.PrimaryEmail) {
			nonPrefixCount++
			continue
		}

		// Don't wait on operations as we may have a lot to delete
		err := usersService.Delete(user.Id).Do()
		if err != nil {
			log.Printf("[INFO][SWEEPER_LOG] Error deleting user: %s", user.PrimaryEmail)
		} else {
			log.Printf("[INFO][SWEEPER_LOG] Sent delete request for %s resource: %s", resourceName, user.PrimaryEmail)
		}
	}

	if nonPrefixCount > 0 {
		log.Printf("[INFO][SWEEPER_LOG] %d items without tf-test prefix remain.", nonPrefixCount)
	}

	return nil
}
