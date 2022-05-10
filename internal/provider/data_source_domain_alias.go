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

type datasourceDomainAliasType struct{}

func (t datasourceDomainAliasType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	resourceType := resourceDomainAliasType{}

	resourceSchema, diags := resourceType.GetSchema(ctx)
	if diags.HasError() {
		return tfsdk.Schema{}, diags
	}

	attrs := datasourceSchemaFromResourceSchema(resourceSchema.Attributes)
	addRequiredFieldsToSchema(attrs, "domain_alias_name")

	return tfsdk.Schema{
		Description: "Domain Alias data source in the Terraform Googleworkspace provider. Domain Alias resides under the " +
			"`https://www.googleapis.com/auth/admin.directory.domain` client scope.",
		Attributes: attrs,
	}, nil
}

type domainAliasDatasource struct {
	provider provider
}

func (t datasourceDomainAliasType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return domainAliasDatasource{
		provider: provider,
	}, diags
}
func (d domainAliasDatasource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data model.DomainAliasResourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainAlias := GetDomainAliasData(&d.provider, &data, &resp.Diagnostics)
	if domainAlias.ID.Value == "" {
		resp.Diagnostics.AddError("object does not exist",
			fmt.Sprintf("Domain Alias %s does not exist", data.DomainAliasName.Value))
	}

	diags = resp.State.Set(ctx, domainAlias)
	resp.Diagnostics.Append(diags...)
}

func GetDomainAliasData(prov *provider, plan *model.DomainAliasResourceData, diags *diag.Diagnostics) *model.DomainAliasResourceData {
	if plan.ID.Null {
		plan.ID = plan.DomainAliasName
	}

	domainAliasesService := GetDomainAliasesService(prov, diags)
	log.Printf("[DEBUG] Getting Domain Alias %s", plan.DomainAliasName.Value)

	domainAliasObj, err := domainAliasesService.Get(prov.customer, plan.DomainAliasName.Value).Do()
	if err != nil {
		plan.ID.Value = handleNotFoundError(err, plan.ID.Value, diags)
	}

	if domainAliasObj == nil {
		diags.AddError("returned obj is nil", fmt.Sprintf("GET %s returned nil object",
			plan.DomainAliasName.Value))
	}

	return SetDomainAliasData(domainAliasObj)
}

func SetDomainAliasData(obj *directory.DomainAlias) *model.DomainAliasResourceData {
	return &model.DomainAliasResourceData{
		ID: types.String{
			Value: obj.DomainAliasName,
		},
		ParentDomainName: types.String{
			Value: obj.ParentDomainName,
		},
		Verified: types.Bool{
			Value: obj.Verified,
		},
		CreationTime: types.Int64{
			Value: obj.CreationTime,
		},
		DomainAliasName: types.String{
			Value: obj.DomainAliasName,
		},
	}
}
