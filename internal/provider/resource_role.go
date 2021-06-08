package googleworkspace

import (
	"context"
	"log"
	"strconv"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	directory "google.golang.org/api/admin/directory/v1"
)

func resourceRole() *schema.Resource {
	return &schema.Resource{
		Description: "Roles resource in the Terraform Googleworkspace provider.",

		CreateContext: resourceRoleCreate,
		ReadContext:   resourceRoleRead,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "ID of the role.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Name of the role.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "A short description of the role.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"privileges": {
				Description: "The set of privileges that are granted to this role.",
				Required:    true,
				Type:        schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_id": {
							Description: "The obfuscated ID of the service this privilege is for.",
							Required:    true,
							Type:        schema.TypeString,
						},
						"privilege_name": {
							Description: "The name of the privilege.",
							Required:    true,
							Type:        schema.TypeString,
						},
					},
				},
			},
			"is_system_role": {
				Description: "Returns true if this is a pre-defined system role.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"is_super_admin_role": {
				Description: "Returns true if the role is a super admin role.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"etag": {
				Description: "ETag of the resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	rolesService, diags := GetRolesService(directoryService)
	if diags.HasError() {
		return diags
	}

	log.Printf("[DEBUG] Creating Role %q", d.Get("name").(string))

	role, err := rolesService.Insert(client.Customer, getRole(d)).Do()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.FormatInt(role.RoleId, 10))

	log.Printf("[DEBUG] Finished creating Role %q", d.Get("name").(string))

	return resourceRoleRead(ctx, d, meta)
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	rolesService, diags := GetRolesService(directoryService)
	if diags.HasError() {
		return diags
	}

	log.Printf("[DEBUG] Updating Role %q", d.Id())

	_, err := rolesService.Update(client.Customer, d.Id(), getRole(d)).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished updating Role %q", d.Id())

	return resourceRoleRead(ctx, d, meta)
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	rolesService, diags := GetRolesService(directoryService)
	if diags.HasError() {
		return diags
	}

	log.Printf("[DEBUG] Getting Role %q", d.Id())

	role, err := rolesService.Get(client.Customer, d.Id()).Do()
	if err != nil {
		return handleNotFoundError(err, d, d.Id())
	}
	if role == nil {
		return diag.Errorf("No Role was returned for %s.", d.Id())
	}

	if diags := setRole(d, role); diags.HasError() {
		return diags
	}

	log.Printf("[DEBUG] Finished getting Role %q", d.Id())

	return diags
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := meta.(*apiClient)

	log.Printf("[DEBUG] Deleting Role %q", d.Id())

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	roleService, diags := GetRolesService(directoryService)
	if diags.HasError() {
		return diags
	}

	err := roleService.Delete(client.Customer, d.Id()).Do()
	if err != nil {
		return handleNotFoundError(err, d, d.Id())
	}

	log.Printf("[DEBUG] Finished deleting Role %q", d.Id())

	return diags
}

func getRole(d *schema.ResourceData) *directory.Role {
	role := &directory.Role{
		RoleName:        d.Get("name").(string),
		RoleDescription: d.Get("description").(string),
	}

	privileges := d.Get("privileges").(*schema.Set)
	for _, pMap := range privileges.List() {
		priv := pMap.(map[string]interface{})
		role.RolePrivileges = append(role.RolePrivileges, &directory.RoleRolePrivileges{
			PrivilegeName: priv["privilege_name"].(string),
			ServiceId:     priv["service_id"].(string),
		})
	}

	if d.Id() != "" {
		id, _ := strconv.ParseInt(d.Id(), 10, 64)
		role.RoleId = id
	}
	return role
}

func setRole(d *schema.ResourceData, role *directory.Role) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(strconv.FormatInt(role.RoleId, 10))
	d.Set("name", role.RoleName)
	d.Set("description", role.RoleDescription)
	d.Set("is_system_role", role.IsSystemRole)
	d.Set("is_super_admin_role", role.IsSuperAdminRole)
	d.Set("etag", role.Etag)

	privileges := make([]interface{}, len(role.RolePrivileges))
	for i, priv := range role.RolePrivileges {
		privileges[i] = map[string]interface{}{
			"service_id":     priv.ServiceId,
			"privilege_name": priv.PrivilegeName,
		}
	}
	if err := d.Set("privileges", privileges); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Error setting attribute",
			Detail:        err.Error(),
			AttributePath: cty.IndexStringPath("privileges"),
		})
	}

	return diags
}
