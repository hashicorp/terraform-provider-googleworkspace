package googleworkspace

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// This will sweep Role Assignmentt resources
func init() {
	resource.AddTestSweepers("RoleAssignment", &resource.Sweeper{
		Name: "RoleAssignment",
		F:    testSweepRoleAssignment,
	})
}

// At the time of writing, the CI only passes us-central1 as the region
func testSweepRoleAssignment(region string) error {
	diags := diag.Diagnostics{}
	resourceName := "Role Assignment"
	log.Printf("[INFO][SWEEPER_LOG] Starting sweeper for %s", resourceName)

	prov := googleworkspaceTestClient(context.Background(), diags)
	if diags.HasError() {
		return fmt.Errorf(getDiagErrors(diags))
	}

	roleAssignmentsService := GetRoleAssignmentsService(prov, &diags)
	if diags.HasError() {
		return fmt.Errorf(getDiagErrors(diags))
	}

	roleAssignments, err := roleAssignmentsService.List(prov.customer).Do()
	if err != nil {
		log.Printf("[INFO][SWEEPER_LOG] Error getting Role Assignments list: %s", err)
		return err
	}

	numRoleAssignments := len(roleAssignments.Items)
	if numRoleAssignments == 0 {
		log.Printf("[INFO][SWEEPER_LOG] Nothing found in response.")
		return nil
	}

	// We need to check the associated role's name to see if we want to sweep the assignments
	usersService := GetUsersService(prov, &diags)
	if diags.HasError() {
		return fmt.Errorf(getDiagErrors(diags))
	}

	log.Printf("[INFO][SWEEPER_LOG] Found %d items in %s list response.", numRoleAssignments, resourceName)
	// Count items that weren't sweeped.
	nonPrefixCount := 0
	for _, roleAssignment := range roleAssignments.Items {
		user, err := usersService.Get(roleAssignment.AssignedTo).Do()
		if err != nil {
			log.Printf("[INFO][SWEEPER_LOG] Error getting associated User: %s", err)
			continue
		}

		if user == nil {
			log.Printf("[INFO][SWEEPER_LOG] Error getting associated User (%s)", roleAssignment.AssignedTo)
			continue
		}

		// Increment count and skip if resource is not sweepable.
		if !isSweepableTestResource(user.PrimaryEmail) {
			log.Printf("[INFO][SWEEPER_LOG] user not sweepable (%s)", user.PrimaryEmail)
			nonPrefixCount++
			continue
		}

		// Don't wait on operations as we may have a lot to delete
		err = roleAssignmentsService.Delete(prov.customer, strconv.FormatInt(roleAssignment.RoleAssignmentId, 10)).Do()
		if err != nil {
			log.Printf("[INFO][SWEEPER_LOG] Error deleting Role Assignment: %s",
				strconv.FormatInt(roleAssignment.RoleAssignmentId, 10))
		} else {
			log.Printf("[INFO][SWEEPER_LOG] Sent delete request for %s resource: %s", resourceName,
				strconv.FormatInt(roleAssignment.RoleAssignmentId, 10))
		}
	}

	if nonPrefixCount > 0 {
		log.Printf("[INFO][SWEEPER_LOG] %d items without tf-test prefix remain.", nonPrefixCount)
	}

	return nil
}
