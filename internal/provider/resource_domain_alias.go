// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	directory "google.golang.org/api/admin/directory/v1"
)

func resourceDomainAlias() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Domain Alias resource manages Google Workspace Domain Aliases. Domain Alias resides under the " +
			"`https://www.googleapis.com/auth/admin.directory.domain` client scope.",

		CreateContext: resourceDomainAliasCreate,
		ReadContext:   resourceDomainAliasRead,
		DeleteContext: resourceDomainAliasDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"parent_domain_name": {
				Description: "The parent domain name that the domain alias is associated with. This can either be a primary or secondary domain name within a customer.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"verified": {
				Description: "Indicates the verification state of a domain alias.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"creation_time": {
				Description: "Creation time of the domain alias.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"etag": {
				Description: "ETag of the resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"domain_alias_name": {
				Description: "The domain alias name.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			// Adding a computed id simply to override the `optional` id that gets added in the SDK
			// that will then display improperly in the docs
			"id": {
				Description: "The ID of this resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceDomainAliasCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	domainAliasName := d.Get("domain_alias_name").(string)
	log.Printf("[DEBUG] Creating DomainAlias %q: %#v", d.Id(), domainAliasName)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	domainAliasesService, diags := GetDomainAliasesService(directoryService)
	if diags.HasError() {
		return diags
	}

	domainAliasObj := directory.DomainAlias{
		ParentDomainName: d.Get("parent_domain_name").(string),
		DomainAliasName:  d.Get("domain_alias_name").(string),
	}

	domainAlias, err := domainAliasesService.Insert(client.Customer, &domainAliasObj).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	// Use the domainAlias name as the ID, as it should be unique
	d.SetId(domainAlias.DomainAliasName)

	log.Printf("[DEBUG] Finished creating DomainAlias %q: %#v", d.Id(), domainAliasName)

	return resourceDomainAliasRead(ctx, d, meta)
}

func resourceDomainAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	domainAliasesService, diags := GetDomainAliasesService(directoryService)
	if diags.HasError() {
		return diags
	}

	log.Printf("[DEBUG] Getting DomainAlias %q: %#v", d.Id(), d.Id())

	domainAlias, err := domainAliasesService.Get(client.Customer, d.Id()).Do()
	if err != nil {
		return handleNotFoundError(err, d, d.Id())
	}

	if domainAlias == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("No domain alias was returned for %s.", d.Get("domain_alias_name").(string)),
		})

		return diags
	}

	d.Set("parent_domain_name", domainAlias.ParentDomainName)
	d.Set("verified", domainAlias.Verified)
	d.Set("creation_time", domainAlias.CreationTime)
	d.Set("etag", domainAlias.Etag)
	d.Set("domain_alias_name", domainAlias.DomainAliasName)
	d.SetId(domainAlias.DomainAliasName)
	log.Printf("[DEBUG] Finished getting DomainAlias %q: %#v", d.Id(), domainAlias.DomainAliasName)

	return diags
}

func resourceDomainAliasDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	domainAliasName := d.Get("domain_alias_name").(string)
	log.Printf("[DEBUG] Deleting DomainAlias %q: %#v", d.Id(), domainAliasName)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	domainAliasesService, diags := GetDomainAliasesService(directoryService)
	if diags.HasError() {
		return diags
	}

	err := domainAliasesService.Delete(client.Customer, domainAliasName).Do()
	if err != nil {
		return handleNotFoundError(err, d, domainAliasName)
	}

	log.Printf("[DEBUG] Finished deleting DomainAlias %q: %#v", d.Id(), domainAliasName)

	return diags
}
