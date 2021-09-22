package googleworkspace

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"reflect"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	directory "google.golang.org/api/admin/directory/v1"
)

func TestAccDataSourcePrivileges_basic(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePrivileges(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.googleworkspace_privileges.test", "etag"),
					resource.TestCheckFunc(testAccResourcePrivilegesCount("data.googleworkspace_privileges.test", "items.#")),
				),
			},
		},
	})
}

func testAccDataSourcePrivileges() string {
	return fmt.Sprintf(`
data "googleworkspace_privileges" "test" {}
`)
}

func testAccResourcePrivilegesCount(resource, attr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("%s key not found in state", resource)
		}

		privCount, err := strconv.Atoi(rs.Primary.Attributes[attr])
		if err != nil {
			return err
		}

		if privCount <= 0 {
			return fmt.Errorf("%s is less than or equal to zero (%d)", attr, privCount)
		}

		client, err := googleworkspaceTestClient()
		if err != nil {
			return err
		}

		directoryService, diags := client.NewDirectoryService()
		if diags.HasError() {
			return fmt.Errorf("error creating directory service %+v", diags)
		}

		privilegesService, diags := GetPrivilegesService(directoryService)
		if diags.HasError() {
			return fmt.Errorf("error getting privileges service %+v", diags)
		}

		privsFromApi, err := privilegesService.List(client.Customer).Do()
		if err != nil {
			return err
		}

		flattenedPrivs := flattenAndPrunePrivileges(privsFromApi.Items, make(map[string]bool))

		if privCount != len(flattenedPrivs) {
			return fmt.Errorf("number of privileges returned (%d) doesn't match number returned by API (%d)", privCount, len(flattenedPrivs))
		}

		return nil
	}
}

func TestDataSourcePrivileges_flattenAndPrune(t *testing.T) {
	t.Parallel()

	input := []*directory.Privilege{
		{
			PrivilegeName: "A",
			ServiceId:     "1",
			ChildPrivileges: []*directory.Privilege{
				{
					PrivilegeName: "AA",
					ServiceId:     "1",
					ChildPrivileges: []*directory.Privilege{
						{
							PrivilegeName: "AAA",
							ServiceId:     "1",
						},
						{ // duplicate
							PrivilegeName: "AAA",
							ServiceId:     "1",
						},
					},
				},
				{
					PrivilegeName: "AB",
					ServiceId:     "1",
				},
			},
		},
		{ // duplicate
			PrivilegeName: "A",
			ServiceId:     "1",
		},
		{
			PrivilegeName: "B",
			ServiceId:     "2",
		},
	}
	expected := []interface{}{
		map[string]interface{}{
			"service_id":           "1",
			"etag":                 "",
			"is_org_unit_scopable": false,
			"privilege_name":       "A",
			"service_name":         "",
		},
		map[string]interface{}{
			"service_id":           "1",
			"etag":                 "",
			"is_org_unit_scopable": false,
			"privilege_name":       "AA",
			"service_name":         "",
		},
		map[string]interface{}{
			"service_id":           "1",
			"etag":                 "",
			"is_org_unit_scopable": false,
			"privilege_name":       "AAA",
			"service_name":         "",
		},
		map[string]interface{}{
			"service_id":           "1",
			"etag":                 "",
			"is_org_unit_scopable": false,
			"privilege_name":       "AB",
			"service_name":         "",
		},
		map[string]interface{}{
			"service_id":           "2",
			"etag":                 "",
			"is_org_unit_scopable": false,
			"privilege_name":       "B",
			"service_name":         "",
		},
	}

	actual := flattenAndPrunePrivileges(input, make(map[string]bool))

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("pruned privilege lists not equal\n\nactual %+v\n\nexpected %+v", actual, expected)
	}
}
