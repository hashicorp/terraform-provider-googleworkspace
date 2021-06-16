package googleworkspace

import (
	"fmt"
	"reflect"
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
					resource.TestCheckResourceAttr("data.googleworkspace_privileges.test", "items.#", "104"),
				),
			},
		},
	})
}

func testAccDataSourcePrivileges() string {
	return fmt.Sprintf(`
data "googleworkspace_privileges" "test" {
}
`)
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
