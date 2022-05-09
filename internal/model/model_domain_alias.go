package model

import "github.com/hashicorp/terraform-plugin-framework/types"

type DomainAliasResourceData struct {
	ID               types.String `tfsdk:"id"`
	ParentDomainName types.String `tfsdk:"parent_domain_name"`
	Verified         types.Bool   `tfsdk:"verified"`
	CreationTime     types.String `tfsdk:"creation_time"`
	DomainAliasName  types.String `tfsdk:"domain_alias_name"`
}
