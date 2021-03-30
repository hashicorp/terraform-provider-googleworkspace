package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	googleoauth "golang.org/x/oauth2/google"
)

var DefaultClientScopes = []string{
	"https://www.googleapis.com/auth/cloud-platform",
	"https://www.googleapis.com/auth/admin.directory.customer",
	"https://www.googleapis.com/auth/admin.directory.domain",
	"https://www.googleapis.com/auth/admin.directory.user",
}

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"credentials": {
					Type:     schema.TypeString,
					Optional: true,
					DefaultFunc: schema.MultiEnvDefaultFunc([]string{
						"GOOGLEWORKSPACE_CREDENTIALS",
						"GOOGLEWORKSPACE_CLOUD_KEYFILE_JSON",
					}, nil),
					ValidateDiagFunc: validateCredentials,
				},

				"customer_id": {
					Type: schema.TypeString,
					DefaultFunc: schema.MultiEnvDefaultFunc([]string{
						"GOOGLEWORKSPACE_CUSTOMER_ID",
					}, nil),
					Optional: true,
				},

				"impersonated_user_email": {
					Type: schema.TypeString,
					DefaultFunc: schema.MultiEnvDefaultFunc([]string{
						"GOOGLEWORKSPACE_IMPERSONATED_USER_EMAIL",
					}, nil),
					Optional: true,
				},

				"oauth_scopes": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"googleworkspace_domain": dataSourceDomain(),
				"googleworkspace_user":   dataSourceUser(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"googleworkspace_domain": resourceDomain(),
				"googleworkspace_user":   resourceUser(),
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

		config.UserAgent = p.UserAgent("terraform-provider-googleworkspace", version)

		diags = config.loadAndValidate(ctx)

		return &config, diags
	}
}

func validateCredentials(v interface{}, p cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if v == nil || v.(string) == "" {
		return diags
	}
	creds := v.(string)
	// if this is a path and we can stat it, assume it's ok
	if _, err := os.Stat(creds); err == nil {
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
