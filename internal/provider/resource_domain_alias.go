package googleworkspace

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-googleworkspace-pf/internal/model"
	"google.golang.org/api/googleapi"
	"log"
)

type resourceDomainAliasType struct{}

// GetSchema Domain Alias Resource
func (r resourceDomainAliasType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Domain Alias resource manages Google Workspace Domain Aliases. Domain Alias resides under the " +
			"`https://www.googleapis.com/auth/admin.directory.domain` client scope.",
		Attributes: map[string]tfsdk.Attribute{
			"parent_domain_name": {
				Description: "The parent domain name that the domain alias is associated with. This can either be a primary or secondary domain name within a customer.",
				Type:        types.StringType,
				Optional:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
			},
			"verified": {
				Description: "Indicates the verification state of a domain alias.",
				Type:        types.BoolType,
				Computed:    true,
			},
			"creation_time": {
				Description: "Creation time of the domain alias.",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"domain_alias_name": {
				Description: "The domain alias name.",
				Type:        types.StringType,
				Required:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "Domain Alias identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

type domainAliasResource struct {
	provider provider
}

func (r resourceDomainAliasType) NewResource(_ context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)

	return domainAliasResource{
		provider: p,
	}, diags
}

// Create a new domain alias
func (r domainAliasResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from "+
				"another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan model.DomainAliasResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainReq := DomainAliasPlanToObj(&plan)

	log.Printf("[DEBUG] Creating Domain Alias %s", plan.DomainAliasName.Value)
	domainAliasesService := GetDomainAliasesService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	domainAliasObj, err := domainAliasesService.Insert(r.provider.customer, &domainReq).Do()
	if err != nil {
		resp.Diagnostics.AddError("error while trying to create domain alias", err.Error())
		return
	}

	if domainAliasObj == nil {
		resp.Diagnostics.AddError(fmt.Sprintf("no domain alias was returned for %s", plan.DomainAliasName.Value), "object returned was nil")
		return
	}
	numInserts := 1

	// INSERT will respond with the Domain Alias that will be created, after INSERT, the etag is updated along with the Domain,
	// once we get a consistent etag, we can feel confident that our Domain is also consistent
	cc := consistencyCheck{
		resourceType: "domain alias",
		timeout:      CreateTimeout,
	}
	err = retryTimeDuration(ctx, CreateTimeout, func() error {
		if cc.reachedConsistency(numInserts) {
			return nil
		}

		newOU, retryErr := domainAliasesService.Get(r.provider.customer, domainAliasObj.DomainAliasName).IfNoneMatch(cc.lastEtag).Do()
		if googleapi.IsNotModified(retryErr) {
			cc.currConsistent += 1
		} else if retryErr != nil {
			return cc.is404(retryErr)
		} else {
			cc.handleNewEtag(newOU.Etag)
		}

		return fmt.Errorf("timed out while waiting for %s to be inserted", cc.resourceType)
	})
	if err != nil {
		return
	}

	plan.ID.Value = domainAliasObj.DomainAliasName
	domainAlias := GetDomainAliasData(&r.provider, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, domainAlias)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Finished creating Domain Alias %s: %s", domainAlias.ID.Value, domainAlias.DomainAliasName.Value)
}

// Read a domain alias
func (r domainAliasResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from "+
				"another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from state
	var state model.DomainAliasResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainAlias := GetDomainAliasData(&r.provider, &state, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &domainAlias)
	resp.Diagnostics.Append(diags...)
}

// Update is not applicable to domain alias
func (r domainAliasResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
}

// Delete a domain alias
func (r domainAliasResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state model.DomainAliasResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Removing Domain Alias %s", state.DomainAliasName.Value)

	resp.State.RemoveResource(ctx)
}

// ImportState a domain alias
func (r domainAliasResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
