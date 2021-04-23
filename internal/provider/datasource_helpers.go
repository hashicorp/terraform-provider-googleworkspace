package googleworkspace

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// datasourceSchemaFromResourceSchema is a recursive func that
// converts an existing Resource schema to a Datasource schema.
// All schema elements are copied, but certain attributes are ignored or changed:
// - all attributes have Computed = true
// - all attributes have ForceNew, Required = false
// - Validation funcs and attributes (e.g. MaxItems) are not copied
func datasourceSchemaFromResourceSchema(rs map[string]*schema.Schema) map[string]*schema.Schema {
	ds := make(map[string]*schema.Schema, len(rs))
	for k, v := range rs {
		dv := &schema.Schema{
			Computed:    true,
			ForceNew:    false,
			Required:    false,
			Description: v.Description,
			Type:        v.Type,
		}

		switch v.Type {
		case schema.TypeSet:
			dv.Set = v.Set
			fallthrough
		case schema.TypeList:
			// List & Set types are generally used for 2 cases:
			// - a list/set of simple primitive values (e.g. list of strings)
			// - a sub resource
			if elem, ok := v.Elem.(*schema.Resource); ok {
				// handle the case where the Element is a sub-resource
				dv.Elem = &schema.Resource{
					Schema: datasourceSchemaFromResourceSchema(elem.Schema),
				}
			} else {
				// handle simple primitive case
				dv.Elem = v.Elem
			}

		default:
			// Elem of all other types are copied as-is
			dv.Elem = v.Elem

		}
		ds[k] = dv

	}
	return ds
}

// addExactlyOneOfFieldsToSchema is a convenience func that sets a list of keys Optional & ExactlyOneOf.
// This is useful when the schema has been generated (using `datasourceSchemaFromResourceSchema` above for
// example) and the datasource could take one multiple inputs (say a unique name or a unique id)
func addExactlyOneOfFieldsToSchema(schema map[string]*schema.Schema, keys ...string) {
	for _, v := range keys {
		schema[v].Computed = false
		schema[v].Optional = true
		schema[v].Required = false
		schema[v].ExactlyOneOf = keys
	}
}
