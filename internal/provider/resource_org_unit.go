package googleworkspace

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	directory "google.golang.org/api/admin/directory/v1"
)

func resourceOrgUnit() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "OrgUnit resource manages Google Workspace OrgUnits. Org Unit resides " +
			"under the `https://www.googleapis.com/auth/admin.directory.orgunit` client scope.",

		CreateContext: resourceOrgUnitCreate,
		ReadContext:   resourceOrgUnitRead,
		UpdateContext: resourceOrgUnitUpdate,
		DeleteContext: resourceOrgUnitDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The organizational unit's path name. For example, an organizational unit's name within the " +
					"/corp/support/sales_support parent path is sales_support.",
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Description: "Description of the organizational unit.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"etag": {
				Description: "ETag of the resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"block_inheritance": {
				Description: "Determines if a sub-organizational unit can inherit the settings of the parent organization. " +
					"False means a sub-organizational unit inherits the settings of the nearest parent organizational unit. " +
					"For more information on inheritance and users in an organization structure, see the " +
					"[administration help center](https://support.google.com/a/answer/4352075).",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"org_unit_id": {
				Description: "The unique ID of the organizational unit.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"org_unit_path": {
				Description: "The full path to the organizational unit. The orgUnitPath is a derived property. " +
					"When listed, it is derived from parentOrgunitPath and organizational unit's name. For example, " +
					"for an organizational unit named 'apps' under parent organization '/engineering', the orgUnitPath " +
					"is '/engineering/apps'. In order to edit an orgUnitPath, either update the name of the organization " +
					"or the parentOrgunitPath. A user's organizational unit determines which Google Workspace services " +
					"the user has access to. If the user is moved to a new organization, the user's access changes. " +
					"For more information about organization structures, see the [administration help center](https://support.google.com/a/answer/4352075). " +
					"For more information about moving a user to a different organization, see " +
					"[chromeosdevices.update a user](https://developers.google.com/admin-sdk/directory/v1/guides/manage-users#update_user).",
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_org_unit_id": {
				Description:  "The unique ID of the parent organizational unit.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"parent_org_unit_id", "parent_org_unit_path"},
			},
			"parent_org_unit_path": {
				Description: "The organizational unit's parent path. For example, /corp/sales is the parent path for " +
					"/corp/sales/sales_support organizational unit.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"parent_org_unit_id", "parent_org_unit_path"},
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

func resourceOrgUnitCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	ouName := d.Get("name").(string)
	log.Printf("[DEBUG] Creating OrgUnit %q: %#v", ouName, ouName)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	orgUnitsService, diags := GetOrgUnitsService(directoryService)
	if diags.HasError() {
		return diags
	}

	orgUnitObj := directory.OrgUnit{
		Name:             d.Get("name").(string),
		Description:      d.Get("description").(string),
		BlockInheritance: d.Get("block_inheritance").(bool),
	}

	if v, ok := d.GetOk("parent_org_unit_id"); ok {
		orgUnitObj.ParentOrgUnitId = v.(string)
	} else {
		orgUnitObj.ParentOrgUnitPath = d.Get("parent_org_unit_path").(string)
	}

	var orgUnit *directory.OrgUnit
	err := retryTimeDuration(ctx, time.Minute, func() error {
		var retryErr error
		orgUnit, retryErr = orgUnitsService.Insert(client.Customer, &orgUnitObj).Do()
		return retryErr
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(orgUnit.OrgUnitId)
	log.Printf("[DEBUG] Finished creating OrgUnit %q: %#v", d.Id(), ouName)

	return resourceOrgUnitRead(ctx, d, meta)
}

func resourceOrgUnitRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	orgUnitsService, diags := GetOrgUnitsService(directoryService)
	if diags.HasError() {
		return diags
	}

	orgUnit, err := orgUnitsService.Get(client.Customer, d.Id()).Do()
	if err != nil {
		return handleNotFoundError(err, d, d.Id())
	}

	d.Set("name", orgUnit.Name)
	d.Set("description", orgUnit.Description)
	d.Set("etag", orgUnit.Etag)
	d.Set("block_inheritance", orgUnit.BlockInheritance)
	d.Set("org_unit_id", orgUnit.OrgUnitId)
	d.Set("org_unit_path", orgUnit.OrgUnitPath)
	d.Set("parent_org_unit_id", orgUnit.ParentOrgUnitId)
	d.Set("parent_org_unit_path", orgUnit.ParentOrgUnitPath)

	d.SetId(orgUnit.OrgUnitId)

	return diags
}

func resourceOrgUnitUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	ouName := d.Get("name").(string)
	log.Printf("[DEBUG] Updating OrgUnit %q: %#v", d.Id(), ouName)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	orgUnitsService, diags := GetOrgUnitsService(directoryService)
	if diags.HasError() {
		return diags
	}

	orgUnitObj := directory.OrgUnit{}

	if d.HasChange("name") {
		orgUnitObj.Name = d.Get("name").(string)
	}

	if d.HasChange("description") {
		orgUnitObj.Description = d.Get("description").(string)
	}

	forceSendFields := []string{}

	if d.HasChange("block_inheritance") {
		orgUnitObj.BlockInheritance = d.Get("block_inheritance").(bool)
		forceSendFields = append(forceSendFields, "BlockInheritance")
	}

	orgUnitObj.ForceSendFields = forceSendFields

	if &orgUnitObj != new(directory.OrgUnit) {
		orgUnit, err := orgUnitsService.Update(client.Customer, d.Id(), &orgUnitObj).Do()
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(orgUnit.OrgUnitId)
	}

	log.Printf("[DEBUG] Finished creating OrgUnit %q: %#v", d.Id(), ouName)

	return resourceOrgUnitRead(ctx, d, meta)
}

func resourceOrgUnitDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	ouName := d.Get("name").(string)
	log.Printf("[DEBUG] Deleting OrgUnit %q: %#v", d.Id(), ouName)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	orgUnitsService, diags := GetOrgUnitsService(directoryService)
	if diags.HasError() {
		return diags
	}

	err := orgUnitsService.Delete(client.Customer, d.Id()).Do()
	if err != nil {
		return handleNotFoundError(err, d, d.Id())
	}

	log.Printf("[DEBUG] Finished deleting OrgUnit %q: %#v", d.Id(), ouName)

	return diags
}
