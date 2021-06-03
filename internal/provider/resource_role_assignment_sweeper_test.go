package googleworkspace

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("RoleAssignment", &resource.Sweeper{
		Name: "RoleAssignment",
		F:    testSweepRoleAssignment,
	})
}

// At the time of writing, the CI only passes us-central1 as the region
func testSweepRoleAssignment(region string) error {
	resourceName := "RoleAssignment"
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

	raService, diags := GetRoleAssignmentsService(directoryService)
	if diags.HasError() {
		log.Printf("[INFO][SWEEPER_LOG] Error getting role assignments service: %s", diags[0].Summary)
		return fmt.Errorf(diags[0].Summary)
	}

	roleAssignments, err := raService.List(client.Customer).Do()
	if err != nil {
		log.Printf("[INFO][SWEEPER_LOG] Error getting role assignments list: %s", err)
		return err
	}

	numRoleAssignments := len(roleAssignments.Items)
	if numRoleAssignments == 0 {
		log.Printf("[INFO][SWEEPER_LOG] Nothing found in response.")
		return nil
	}

	log.Printf("[INFO][SWEEPER_LOG] Found %d items in %s list response.", numRoleAssignments, resourceName)

	for _, ra := range roleAssignments.Items {
		err := raService.Delete(client.Customer, strconv.FormatInt(ra.RoleAssignmentId, 10)).Do()
		if err != nil {
			log.Printf("[INFO][SWEEPER_LOG] Error deleting role assignment: %v", ra.RoleAssignmentId)
		} else {
			log.Printf("[INFO][SWEEPER_LOG] Sent delete request for %s resource: %v", resourceName, ra.RoleAssignmentId)
		}
	}

	return nil
}
