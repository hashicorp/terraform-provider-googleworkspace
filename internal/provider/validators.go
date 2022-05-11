package googleworkspace

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strings"
)

type ExactlyOneOfValidator struct {
	RequiredAttrs []string
}

func (v ExactlyOneOfValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {

}

func (v ExactlyOneOfValidator) Description(ctx context.Context) string {
	return "Validates that one of the supplied attributes is set in the user's config."
}

func (v ExactlyOneOfValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates that one of the supplied attributes is set in the user's config."
}

// StringLenBetweenValidator validates that the length of the supplied string is between Min and Max (inclusive)
type StringLenBetweenValidator struct {
	Min int
	Max int
}

func (v StringLenBetweenValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	attrName := req.AttributePath.String()
	attrValue := req.AttributeConfig.(types.String).Value

	if len(attrValue) < v.Min {
		resp.Diagnostics.AddError("length of string is not long enough", fmt.Sprintf("length of %s (%s) needs at least %d characters", attrName, attrValue, v.Min))
	}

	if len(attrValue) > v.Max {
		resp.Diagnostics.AddError("length of string is too long", fmt.Sprintf("length of %s (%s) needs at less than %d characters", attrName, attrValue, v.Max))
	}
}

func (v StringLenBetweenValidator) Description(ctx context.Context) string {
	return "Validates that the length of the supplied string is between Min and Max (inclusive)."
}

func (v StringLenBetweenValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates that the length of the supplied string is between Min and Max (inclusive)."
}

// StringInSliceValidator validates that the provided string is a valid option
type StringInSliceValidator struct {
	Options []string
}

func (v StringInSliceValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	attrName := req.AttributePath.String()
	attrValue := req.AttributeConfig.(types.String).Value

	if !stringInSlice(v.Options, attrValue) {
		resp.Diagnostics.AddError("string value is not a valid option", fmt.Sprintf("%s (%s) must be one of [%s]", attrName, attrValue, strings.Join(v.Options, ", ")))
	}
}

func (v StringInSliceValidator) Description(ctx context.Context) string {
	return "Validates that the provided string is a valid option."
}

func (v StringInSliceValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates that the provided string is a valid option."
}
