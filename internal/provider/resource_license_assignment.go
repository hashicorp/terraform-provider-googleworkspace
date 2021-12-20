package googleworkspace

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	licensing "google.golang.org/api/licensing/v1"
)

func resourceLicenseAssignment() *schema.Resource {
	return &schema.Resource{
		Description: "License Assignment Resource",
		CreateContext: resourceLicenseAssignmentCreate,
		ReadContext:   resourceLicenseAssignmentRead,
		UpdateContext: resourceLicenseAssignmentUpdate,
		DeleteContext: resourceLicenseAssignmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
		    "user_email": {
            	Description: "The user's current primary email address.",
            	Type:        schema.TypeString,
            	Required:    true,
            },
            "sku_id": {
                Description: "A product SKU's unique identifier.",
                Type:        schema.TypeString,
                Required:    true,
            },
            "product_id": {
                 Description: "A product's unique identifier.",
                 Type:        schema.TypeString,
                 Required:    true,
                 ForceNew:    true,
            },
		},
	}
}

func resourceLicenseAssignmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := meta.(*apiClient)

	productId := d.Get("product_id").(string)
	skuId := d.Get("sku_id").(string)
	userId := d.Get("user_email").(string)

	licensingService, diags := client.NewLicensingService()
	if diags.HasError() {
		return diags
	}

	licenseAssignmentsService, diags := GetLicenseAssignmentsService(licensingService)
	if diags.HasError() {
		return diags
	}

	la := &licensing.LicenseAssignmentInsert{
		UserId: userId,
	}

	_, err := licenseAssignmentsService.Insert(productId, skuId, la).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(productId + "/" + skuId + "/" + userId)

	return resourceLicenseAssignmentRead(ctx, d, meta)
}

func resourceLicenseAssignmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := meta.(*apiClient)

    parts := strings.Split(d.Id(), "/")

    licensingService, diags := client.NewLicensingService()
    if diags.HasError() {
       return diags
    }

    licenseAssignmentsService, diags := GetLicenseAssignmentsService(licensingService)
    if diags.HasError() {
    	return diags
    }

    la, err := licenseAssignmentsService.Get(parts[0], parts[1], parts[2]).Do()
    if err != nil {
    	return diag.FromErr(err)
    }

    d.Set("user_email", la.UserId)
    d.Set("sku_id", la.SkuId)
    d.Set("product_id", la.ProductId)
    d.SetId(la.ProductId + "/" + la.SkuId + "/" + la.UserId)

    return diags
}

func resourceLicenseAssignmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := meta.(*apiClient)

    productId := d.Get("product_id").(string)
    skuId := d.Get("sku_id").(string)
    userId := d.Get("user_email").(string)

    licensingService, diags := client.NewLicensingService()
    if diags.HasError() {
       return diags
    }

    licenseAssignmentsService, diags := GetLicenseAssignmentsService(licensingService)
    if diags.HasError() {
    	return diags
    }

    la := &licensing.LicenseAssignment{
        UserId: userId,
        ProductId: productId,
        SkuId: skuId,
    }

    if d.HasChange("user_email") {
        d.SetId(la.ProductId + "/" + la.SkuId + "/" + la.UserId)
    }

    parts := strings.Split(d.Id(), "/")

    _, err := licenseAssignmentsService.Update(parts[0], parts[1], parts[2], la).Do()
    if err != nil {
    	return diag.FromErr(err)
    }

    if d.HasChange("sku_id") {
        d.SetId(la.ProductId + "/" + la.SkuId + "/" + la.UserId)
    }

    return resourceLicenseAssignmentRead(ctx, d, meta)
}

func resourceLicenseAssignmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    var diags diag.Diagnostics
    client := meta.(*apiClient)

    parts := strings.Split(d.Id(), "/")

    licensingService, diags := client.NewLicensingService()
    if diags.HasError() {
       return diags
    }

    licenseAssignmentsService, diags := GetLicenseAssignmentsService(licensingService)
    if diags.HasError() {
    	return diags
    }

    _, err := licenseAssignmentsService.Delete(parts[0], parts[1], parts[2]).Do()
    if err != nil {
    	return diag.FromErr(err)
    }

    d.SetId("")

    return diags
}

