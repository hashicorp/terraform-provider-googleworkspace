package model

import "github.com/hashicorp/terraform-plugin-framework/types"

type SchemaResourceData struct {
	ID          types.String `tfsdk:"id"`
	SchemaId    types.String `tfsdk:"schema_id"`
	SchemaName  types.String `tfsdk:"schema_name"`
	DisplayName types.String `tfsdk:"display_name"`
	Fields      types.List   `tfsdk:"fields"`
}

type SchemaResourceField struct {
	FieldName           types.String `tfsdk:"field_name"`
	FieldId             types.String `tfsdk:"field_id"`
	FieldType           types.String `tfsdk:"field_type"`
	MultiValued         types.Bool   `tfsdk:"multi_valued"`
	Indexed             types.Bool   `tfsdk:"indexed"`
	DisplayName         types.String `tfsdk:"display_name"`
	ReadAccessType      types.String `tfsdk:"read_access_type"`
	NumericIndexingSpec types.Object `tfsdk:"numeric_indexing_spec"`
	Fields              types.List   `tfsdk:"fields"`
}

type SchemaResourceFieldNumericIndexingSpec struct {
	MinValue types.Float64 `tfsdk:"min_value"`
	MaxValue types.Float64 `tfsdk:"max_value"`
}
