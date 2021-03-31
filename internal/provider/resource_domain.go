package googleworkspace

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	directory "google.golang.org/api/admin/directory/v1"
)

func resourceDomain() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Domain resource manages Google Workspace Domains.",

		CreateContext: resourceDomainCreate,
		ReadContext:   resourceDomainRead,
		DeleteContext: resourceDomainDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceDomainImport,
		},

		Schema: map[string]*schema.Schema{
			"domain_aliases": {
				Description: "asps.list of domain alias objects.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"verified": {
				Description: "Indicates the verification state of a domain.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"etag": {
				Description: "ETag of the resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"creation_time": {
				Description: "Creation time of the domain. Expressed in Unix time format.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"is_primary": {
				Description: "Indicates if the domain is a primary domain.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"domain_name": {
				Description: "The domain name of the customer.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	domainName := d.Get("domain_name").(string)
	log.Printf("[DEBUG] Creating Domain %q: %#v", d.Id(), domainName)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	domainsService, diags := GetDomainsService(directoryService)
	if diags.HasError() {
		return diags
	}

	domainObj := directory.Domains{
		DomainName: d.Get("domain_name").(string),
	}

	domain, err := domainsService.Insert(client.Customer, &domainObj).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	// Use the domain name as the ID, as it should be unique
	d.SetId(domain.DomainName)

	log.Printf("[DEBUG] Finished creating Domain %q: %#v", d.Id(), domainName)

	return resourceDomainRead(ctx, d, meta)
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	domainsService, diags := GetDomainsService(directoryService)
	if diags.HasError() {
		return diags
	}

	domainName := d.Get("domain_name").(string)

	domain, err := domainsService.Get(client.Customer, domainName).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("domain_aliases", flattenDomainAliases(domain.DomainAliases, d)); err != nil {
		return diag.FromErr(err)
	}

	d.Set("verified", domain.Verified)
	d.Set("creation_time", domain.CreationTime)
	d.Set("is_primary", domain.IsPrimary)
	d.Set("domain_name", domain.DomainName)
	d.SetId(domain.DomainName)

	return diags
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	domainName := d.Get("domain_name").(string)
	log.Printf("[DEBUG] Deleting Domain %q: %#v", d.Id(), domainName)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	domainsService, diags := GetDomainsService(directoryService)
	if diags.HasError() {
		return diags
	}

	err := domainsService.Delete(client.Customer, domainName).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished deleting Domain %q: %#v", d.Id(), domainName)

	return diags
}

func resourceDomainImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	err := d.Set("domain_name", d.Id())

	return []*schema.ResourceData{d}, err
}

func flattenDomainAliases(domainAliases []*directory.DomainAlias, d *schema.ResourceData) interface{} {
	var v []string

	for _, domainAlias := range domainAliases {
		v = append(v, domainAlias.DomainAliasName)
	}

	return v
}
