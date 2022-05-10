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

type resourceDomainType struct{}

// GetSchema Domain Resource
func (r resourceDomainType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "Domain resource manages Google Workspace Domains. Domain resides under the " +
			"`https://www.googleapis.com/auth/admin.directory.domain` client scope.",
		Attributes: map[string]tfsdk.Attribute{
			"domain_aliases": {
				Description: "asps.list of domain alias objects.",
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Computed: true,
			},
			"verified": {
				Description: "Indicates the verification state of a domain.",
				Type:        types.BoolType,
				Computed:    true,
			},
			"creation_time": {
				Description: "Creation time of the domain. Expressed in Unix time format.",
				Type:        types.Int64Type,
				Computed:    true,
			},
			"is_primary": {
				Description: "Indicates if the domain is a primary domain.",
				Type:        types.BoolType,
				Computed:    true,
			},
			"domain_name": {
				Description: "The domain name of the customer.",
				Type:        types.StringType,
				Required:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "Domain identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

type domainResource struct {
	provider provider
}

func (r resourceDomainType) NewResource(_ context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)

	return domainResource{
		provider: p,
	}, diags
}

// Create a new domain
func (r domainResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from "+
				"another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan model.DomainResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainReq := DomainPlanToObj(&plan)

	log.Printf("[DEBUG] Creating Domain %s", plan.DomainName.Value)
	domainsService := GetDomainsService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	domainObj, err := domainsService.Insert(r.provider.customer, &domainReq).Do()
	if err != nil {
		resp.Diagnostics.AddError("error while trying to create domain", err.Error())
		return
	}

	if domainObj == nil {
		resp.Diagnostics.AddError(fmt.Sprintf("no domain was returned for %s", plan.DomainName.Value), "object returned was nil")
		return
	}
	numInserts := 1

	// INSERT will respond with the Domain that will be created, after INSERT, the etag is updated along with the Domain,
	// once we get a consistent etag, we can feel confident that our Domain is also consistent
	cc := consistencyCheck{
		resourceType: "domain",
		timeout:      CreateTimeout,
	}
	err = retryTimeDuration(ctx, CreateTimeout, func() error {
		if cc.reachedConsistency(numInserts) {
			return nil
		}

		newOU, retryErr := domainsService.Get(r.provider.customer, domainObj.DomainName).IfNoneMatch(cc.lastEtag).Do()
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

	plan.ID.Value = domainObj.DomainName
	domain := GetDomainData(&r.provider, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, domain)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Finished creating Domain %s: %s", domain.ID.Value, domain.DomainName.Value)
}

// Read a domain
func (r domainResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from "+
				"another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from state
	var state model.DomainResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := GetDomainData(&r.provider, &state, &resp.Diagnostics)

	diags = resp.State.Set(ctx, &domain)
	resp.Diagnostics.Append(diags...)
}

// Update is not applicable to domain
func (r domainResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
}

// Delete a domain
func (r domainResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state model.DomainResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Removing Domain %s", state.DomainName.Value)

	resp.State.RemoveResource(ctx)
}

// ImportState a domain
func (r domainResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}
