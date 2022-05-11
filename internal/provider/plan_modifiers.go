package googleworkspace

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DefaultModifier allows an attribute to supply a default value if the previously planned value is null
type DefaultModifier struct {
	DefaultValue attr.Value
}

func (m DefaultModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if req.AttributeConfig == nil || req.AttributePlan == nil || req.AttributeState == nil {
		// shouldn't happen, but let's not panic if it does
		return
	}

	attrType := req.AttributePlan.Type(ctx)

	switch attrType {
	case types.StringType:
		if req.AttributePlan.(types.String).Null {
			resp.AttributePlan = m.DefaultValue
			return
		}
	}
}

// Description returns a human-readable description of the plan modifier.
func (m DefaultModifier) Description(ctx context.Context) string {
	return "If the value of this attribute is null, the given value will be used instead."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m DefaultModifier) MarkdownDescription(ctx context.Context) string {
	return "If the value of this attribute is null, the given value will be used instead."
}
