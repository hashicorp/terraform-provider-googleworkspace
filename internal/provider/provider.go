package googleworkspace

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/go-homedir"

	googleoauth "golang.org/x/oauth2/google"
)

var DefaultClientScopes = []string{
	"https://www.googleapis.com/auth/gmail.settings.basic",
	"https://www.googleapis.com/auth/gmail.settings.sharing",
	"https://www.googleapis.com/auth/chrome.management.policy",
	"https://www.googleapis.com/auth/cloud-platform",
	"https://www.googleapis.com/auth/admin.directory.customer",
	"https://www.googleapis.com/auth/admin.directory.domain",
	"https://www.googleapis.com/auth/admin.directory.group",
	"https://www.googleapis.com/auth/admin.directory.orgunit",
	"https://www.googleapis.com/auth/admin.directory.rolemanagement",
	"https://www.googleapis.com/auth/admin.directory.userschema",
	"https://www.googleapis.com/auth/admin.directory.user",
	"https://www.googleapis.com/auth/apps.groups.settings",
}

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		desc := s.Description
		if s.Default != nil {
			desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
		}
		return strings.TrimSpace(desc)
	}
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"access_token": {
					Description: "A temporary [OAuth 2.0 access token] obtained from " +
						"the Google Authorization server, i.e. the `Authorization: Bearer` token used to " +
						"authenticate HTTP requests to Google Admin SDK APIs. This is an alternative to `credentials`, " +
						"and ignores the `scopes` field. If both are specified, `access_token` will be " +
						"used over the `credentials` field.",
					Type:     schema.TypeString,
					Optional: true,
				},

				"credentials": {
					Description: "Either the path to or the contents of a service account key file in JSON format " +
						"you can manage key files using the Cloud Console).  If not provided, the application default " +
						"credentials will be used.",
					Type:     schema.TypeString,
					Optional: true,
					DefaultFunc: schema.MultiEnvDefaultFunc([]string{
						"GOOGLEWORKSPACE_CREDENTIALS",
						"GOOGLEWORKSPACE_CLOUD_KEYFILE_JSON",
						"GOOGLE_CREDENTIALS",
					}, nil),
					ValidateDiagFunc: validateCredentials,
				},

				"customer_id": {
					Description: "The customer id provided with your Google Workspace subscription. It is found " +
						"in the admin console under Account Settings.",
					Type: schema.TypeString,
					DefaultFunc: schema.MultiEnvDefaultFunc([]string{
						"GOOGLEWORKSPACE_CUSTOMER_ID",
					}, nil),
					Optional: true,
				},

				"impersonated_user_email": {
					Description: "The impersonated user's email with access to the Admin APIs can access the Admin SDK Directory API. " +
						"`impersonated_user_email` is required for all services except group and user management.",
					Type: schema.TypeString,
					DefaultFunc: schema.MultiEnvDefaultFunc([]string{
						"GOOGLEWORKSPACE_IMPERSONATED_USER_EMAIL",
					}, nil),
					Optional: true,
				},

				"oauth_scopes": {
					Description: "The list of the scopes required for your application (for a list of possible scopes, see " +
						"[Authorize requests](https://developers.google.com/admin-sdk/directory/v1/guides/authorizing))",
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},

				"service_account": {
					Description: "The service account used to create the provided `access_token` if authenticating using " +
						"the `access_token` method and needing to impersonate a user. This service account will require the " +
						"GCP role `Service Account Token Creator` if needing to impersonate a user.",
					Type:     schema.TypeString,
					Optional: true,
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"googleworkspace_chrome_policy_schema": dataSourceChromePolicySchema(),
				"googleworkspace_domain":               dataSourceDomain(),
				"googleworkspace_domain_alias":         dataSourceDomainAlias(),
				"googleworkspace_group":                dataSourceGroup(),
				"googleworkspace_group_member":         dataSourceGroupMember(),
				"googleworkspace_group_members":        dataSourceGroupMembers(),
				"googleworkspace_group_settings":       dataSourceGroupSettings(),
				"googleworkspace_org_unit":             dataSourceOrgUnit(),
				"googleworkspace_privileges":           dataSourcePrivileges(),
				"googleworkspace_role":                 dataSourceRole(),
				"googleworkspace_schema":               dataSourceSchema(),
				"googleworkspace_user":                 dataSourceUser(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"googleworkspace_chrome_policy":       resourceChromePolicy(),
				"googleworkspace_domain":              resourceDomain(),
				"googleworkspace_domain_alias":        resourceDomainAlias(),
				"googleworkspace_gmail_send_as_alias": resourceGmailSendAsAlias(),
				"googleworkspace_group":               resourceGroup(),
				"googleworkspace_group_member":        resourceGroupMember(),
				"googleworkspace_group_members":       resourceGroupMembers(),
				"googleworkspace_group_settings":      resourceGroupSettings(),
				"googleworkspace_org_unit":            resourceOrgUnit(),
				"googleworkspace_role":                resourceRole(),
				"googleworkspace_role_assignment":     resourceRoleAssignment(),
				"googleworkspace_schema":              resourceSchema(),
				"googleworkspace_user":                resourceUser(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		var diags diag.Diagnostics
		config := apiClient{}

		// Get access token
		if v, ok := d.GetOk("access_token"); ok {
			config.AccessToken = v.(string)
		}

		// Get credentials
		if v, ok := d.GetOk("credentials"); ok {
			config.Credentials = v.(string)
		}

		// Get customer id
		if v, ok := d.GetOk("customer_id"); ok {
			config.Customer = v.(string)
		} else {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "customer_id is required",
			})

			return nil, diags
		}

		// Get impersonated user email
		if v, ok := d.GetOk("impersonated_user_email"); ok {
			config.ImpersonatedUserEmail = v.(string)
		}

		// Get scopes
		scopes := d.Get("oauth_scopes").([]interface{})
		if len(scopes) > 0 {
			config.ClientScopes = make([]string, len(scopes))
		}
		for i, scope := range scopes {
			config.ClientScopes[i] = scope.(string)
		}

		// Get service account
		if v, ok := d.GetOk("service_account"); ok {
			config.ServiceAccount = v.(string)
		}

		config.UserAgent = p.UserAgent("terraform-provider-googleworkspace", version)

		newCtx, _ := schema.StopContext(ctx)
		diags = config.loadAndValidate(newCtx)

		return &config, diags
	}
}

func validateCredentials(v interface{}, p cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if v == nil || v.(string) == "" {
		return diags
	}
	creds := v.(string)
	path, err := homedir.Expand(creds)
	if err != nil {
		return diag.FromErr(err)
	}
	// if this is a path and we can stat it, assume it's ok
	if _, err := os.Stat(path); err == nil {
		return diags
	}
	if _, err := googleoauth.CredentialsFromJSON(context.Background(), []byte(creds)); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       fmt.Sprintf("JSON credentials in %q are not valid: %s", creds, err),
			AttributePath: p,
		})
	}

	return diags
}
