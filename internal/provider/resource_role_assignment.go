package googleworkspace

import (
	"context"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	directory "google.golang.org/api/admin/directory/v1"
)

func resourceRoleAssignment() *schema.Resource {
	return &schema.Resource{
		Description:   "RoleAssignment resource in the Terraform Googleworkspace provider.",
		CreateContext: resourceRolesAssignmentCreate,
		ReadContext:   resourceRoleAssignmentRead,
		DeleteContext: resourceRoleAssignmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "ID of this roleAssignment.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"role_id": {
				Description: "The ID of the role that is assigned.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"etag": {
				Description: "ETag of the resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"assigned_to": {
				Description: "The unique ID of the user this role is assigned to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			/* TODO: ORG_UNIT scope not working at this time
			"scope_type": {
				Description:      "The scope in which this role is assigned.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"CUSTOMER", "ORG_UNIT"}, false)),
				ForceNew:         true,
			},
			"org_unit_id": {
				Description: "If the role is restricted to an organization unit, this contains the ID for the organization unit the exercise of this role is restricted to.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			*/
		},
	}
}

func resourceRolesAssignmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := meta.(*apiClient)

	assignedTo := d.Get("assigned_to").(string)
	roleId := d.Get("role_id").(string)
	log.Printf("[DEBUG] Creating RoleAssignment user:%s, role:%s", assignedTo, roleId)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	roleAssignmentsService, diags := GetRoleAssignmentsService(directoryService)
	if diags.HasError() {
		return diags
	}

	roleIdInt64, err := strconv.ParseInt(roleId, 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}
	ra := &directory.RoleAssignment{
		AssignedTo: assignedTo,
		RoleId:     roleIdInt64,
		ScopeType:  "CUSTOMER", // hardcoded until ORG_UNIT scope is sorted out
	}

	ra, err = roleAssignmentsService.Insert(client.Customer, ra).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(ra.RoleAssignmentId, 10))

	log.Printf("[DEBUG] Finished creating RoleAssignment user:%s, role:%s", assignedTo, roleId)

	return resourceRoleAssignmentRead(ctx, d, meta)
}

func resourceRoleAssignmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := meta.(*apiClient)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	roleAssignmentsService, diags := GetRoleAssignmentsService(directoryService)
	if diags.HasError() {
		return diags
	}

	log.Printf("[DEBUG] Getting RoleAssignment %q", d.Id())

	ra, err := roleAssignmentsService.Get(client.Customer, d.Id()).Do()
	if err != nil {
		return handleNotFoundError(err, d, d.Id())
	}
	if ra == nil {
		return diag.Errorf("No RoleAssignment was returned for %s.", d.Id())
	}

	d.SetId(strconv.FormatInt(ra.RoleAssignmentId, 10))
	d.Set("role_id", strconv.FormatInt(ra.RoleId, 10))
	d.Set("etag", ra.Etag)
	d.Set("assigned_to", ra.AssignedTo)

	log.Printf("[DEBUG] Finished getting RoleAssignment %q", d.Id())

	return diags
}

func resourceRoleAssignmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := meta.(*apiClient)

	log.Printf("[DEBUG] Deleting RoleAssignment %q", d.Id())

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	roleAssignmentsService, diags := GetRoleAssignmentsService(directoryService)
	if diags.HasError() {
		return diags
	}

	err := roleAssignmentsService.Delete(client.Customer, d.Id()).Do()
	if err != nil {
		return handleNotFoundError(err, d, d.Id())
	}

	log.Printf("[DEBUG] Finished deleting RoleAssignment %q", d.Id())

	return diags
}
