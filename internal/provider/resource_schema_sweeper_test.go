package googleworkspace

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// This will sweep Schema resources
func init() {
	resource.AddTestSweepers("Schema", &resource.Sweeper{
		Name: "Schema",
		F:    testSweepSchema,
	})
}

// At the time of writing, the CI only passes us-central1 as the region
func testSweepSchema(region string) error {
	resourceName := "Schema"
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

	schemasService, diags := GetSchemasService(directoryService)
	if diags.HasError() {
		log.Printf("[INFO][SWEEPER_LOG] Error getting schemas service: %s", diags[0].Summary)
		return fmt.Errorf(diags[0].Summary)
	}

	schemas, err := schemasService.List(client.Customer).Do()
	if err != nil {
		log.Printf("[INFO][SWEEPER_LOG] Error getting schema list: %s", err)
		return err
	}

	numSchemas := len(schemas.Schemas)
	if numSchemas == 0 {
		log.Printf("[INFO][SWEEPER_LOG] Nothing found in response.")
		return nil
	}

	log.Printf("[INFO][SWEEPER_LOG] Found %d items in %s list response.", numSchemas, resourceName)
	// Count items that weren't sweeped.
	nonPrefixCount := 0
	for _, schema := range schemas.Schemas {
		// Increment count and skip if resource is not sweepable.
		if !isSweepableTestResource(schema.SchemaName) {
			nonPrefixCount++
			continue
		}

		// Don't wait on operations as we may have a lot to delete
		err := schemasService.Delete(client.Customer, schema.SchemaId).Do()
		if err != nil {
			log.Printf("[INFO][SWEEPER_LOG] Error deleting schema: %s", schema.SchemaName)
		} else {
			log.Printf("[INFO][SWEEPER_LOG] Sent delete request for %s resource: %s", resourceName, schema.SchemaName)
		}
	}

	if nonPrefixCount > 0 {
		log.Printf("[INFO][SWEEPER_LOG] %d items without tf-test prefix remain.", nonPrefixCount)
	}

	return nil
}
