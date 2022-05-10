package googleworkspace

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-provider-googleworkspace-pf/internal/model"
	directory "google.golang.org/api/admin/directory/v1"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type datasourceDomainType struct{}

func (t datasourceDomainType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	resourceType := resourceDomainType{}

	resourceSchema, diags := resourceType.GetSchema(ctx)
	if diags.HasError() {
		return tfsdk.Schema{}, diags
	}

	attrs := datasourceSchemaFromResourceSchema(resourceSchema.Attributes)
	addRequiredFieldsToSchema(attrs, "domain_name")

	return tfsdk.Schema{
		Description: "Domain data source in the Terraform Googleworkspace provider. Domain resides under the " +
			"`https://www.googleapis.com/auth/admin.directory.domain` client scope.",
		Attributes: attrs,
	}, nil
}

type domainDatasource struct {
	provider provider
}

func (t datasourceDomainType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	p, diags := convertProviderType(in)

	return domainDatasource{
		provider: p,
	}, diags
}

func (d domainDatasource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data model.DomainResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := GetDomainData(&d.provider, &data, &resp.Diagnostics)
	if domain.ID.Value == "" {
		resp.Diagnostics.AddError("object does not exist",
			fmt.Sprintf("Domain %s does not exist", data.DomainName.Value))
	}

	diags = resp.State.Set(ctx, domain)
	resp.Diagnostics.Append(diags...)
}

func GetDomainData(prov *provider, plan *model.DomainResourceData, diags *diag.Diagnostics) *model.DomainResourceData {
	if plan.ID.Null {
		plan.ID = plan.DomainName
	}

	domainsService := GetDomainsService(prov, diags)
	log.Printf("[DEBUG] Getting Domain %s", plan.DomainName.Value)

	domainObj, err := domainsService.Get(prov.customer, plan.DomainName.Value).Do()
	if err != nil {
		plan.ID.Value = handleNotFoundError(err, plan.ID.Value, diags)
	}

	if domainObj == nil {
		diags.AddError("returned obj is nil", fmt.Sprintf("GET %s returned nil object",
			plan.DomainName.Value))
	}

	return SetDomainData(domainObj)
}

func SetDomainData(obj *directory.Domains) *model.DomainResourceData {
	var domainAliases types.List
	for _, alias := range obj.DomainAliases {
		domainAliases.Elems = append(domainAliases.Elems, types.String{Value: alias.DomainAliasName})
	}

	return &model.DomainResourceData{
		ID: types.String{
			Value: obj.DomainName,
		},
		DomainAliases: domainAliases,
		Verified: types.Bool{
			Value: obj.Verified,
		},
		CreationTime: types.Int64{
			Value: obj.CreationTime,
		},
		IsPrimary: types.Bool{
			Value: obj.IsPrimary,
		},
		DomainName: types.String{
			Value: obj.DomainName,
		},
	}
}
