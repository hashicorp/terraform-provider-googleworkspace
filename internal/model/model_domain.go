package model

import "github.com/hashicorp/terraform-plugin-framework/types"

type DomainResourceData struct {
	ID            types.String `tfsdk:"id"`
	DomainAliases types.List   `tfsdk:"domain_aliases"`
	Verified      types.Bool   `tfsdk:"verified"`
	CreationTime  types.Int64  `tfsdk:"creation_time"`
	IsPrimary     types.Bool   `tfsdk:"is_primary"`
	DomainName    types.String `tfsdk:"domain_name"`
}
