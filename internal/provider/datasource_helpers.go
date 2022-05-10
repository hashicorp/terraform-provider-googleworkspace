package googleworkspace

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// datasourceSchemaFromResourceSchema is a recursive func that
// converts an existing Resource schema to a Datasource schema.
// All schema elements are copied, but certain attributes are ignored or changed:
// - all attributes have Computed = true
// - all attributes have ForceNew, Required = false
// - Validation funcs and plan modifiers are not copied
func datasourceSchemaFromResourceSchema(rs map[string]tfsdk.Attribute) map[string]tfsdk.Attribute {
	ds := make(map[string]tfsdk.Attribute, len(rs))
	for k, v := range rs {
		dv := tfsdk.Attribute{
			Computed:    true,
			Required:    false,
			Description: v.Description,
			Type:        v.Type,
		}

		switch v.Type {
		case types.SetType{}:
			dv.Attributes = tfsdk.SetNestedAttributes(datasourceSchemaFromResourceSchema(v.Attributes.GetAttributes()), tfsdk.SetNestedAttributesOptions{})
		case types.ListType{}:
			dv.Attributes = tfsdk.ListNestedAttributes(datasourceSchemaFromResourceSchema(v.Attributes.GetAttributes()), tfsdk.ListNestedAttributesOptions{})
		case types.ObjectType{}:
			dv.Attributes = tfsdk.SingleNestedAttributes(datasourceSchemaFromResourceSchema(v.Attributes.GetAttributes()))
		case types.MapType{}:
			dv.Attributes = tfsdk.MapNestedAttributes(datasourceSchemaFromResourceSchema(v.Attributes.GetAttributes()), tfsdk.MapNestedAttributesOptions{})
		default:
			// Elem of all other types are copied as-is
			dv.Attributes = v.Attributes
		}
		ds[k] = dv

	}
	return ds
}

// fixDatasourceSchemaFlags is a convenience func that toggles the Computed,
// Optional + Required flags on a schema element. This is useful when the schema
// has been generated (using `datasourceSchemaFromResourceSchema` above for
// example) and therefore the attribute flags were not set appropriately when
// first added to the schema definition. Currently only supports top-level
// schema elements.
func fixDatasourceSchemaFlags(schema map[string]tfsdk.Attribute, required bool, keys ...string) {
	for _, v := range keys {
		attr := schema[v]
		(&attr).Computed = false
		(&attr).Optional = !required
		(&attr).Required = required
	}
}

func addRequiredFieldsToSchema(schema map[string]tfsdk.Attribute, keys ...string) {
	fixDatasourceSchemaFlags(schema, true, keys...)
}

// addExactlyOneOfFieldsToSchema is a convenience func that sets a list of keys Optional & ExactlyOneOf.
// This is useful when the schema has been generated (using `datasourceSchemaFromResourceSchema` above for
// example) and the datasource could take one multiple inputs (say a unique name or a unique id)
func addExactlyOneOfFieldsToSchema(schema map[string]tfsdk.Attribute, keys ...string) {
	for _, v := range keys {
		attr := schema[v]
		(&attr).Computed = true
		(&attr).Optional = true
		(&attr).Required = false
		//(&attr).ExactlyOneOf = keys
	}
}
