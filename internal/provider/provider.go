package googleworkspace

import (
	"context"
	"fmt"

	//"fmt"
	"net/http"
	"os"
	//"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	//"github.com/mitchellh/go-homedir"
	//googleoauth "golang.org/x/oauth2/google"
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

var CredentialsEnvVars = []string{
	"GOOGLEWORKSPACE_CREDENTIALS",
	"GOOGLEWORKSPACE_CLOUD_KEYFILE_JSON",
	"GOOGLE_CREDENTIALS",
}

var stderr = os.Stderr

func New(version string) func() tfsdk.Provider {
	return func() tfsdk.Provider {
		return &provider{
			version: version,
		}
	}
}

type provider struct {
	client      *http.Client
	configured  bool
	credentials string
	customer    string
	oauthscopes types.List
	version     string
	userAgent   string
}

type providerData struct {
	AccessToken           types.String `tfsdk:"access_token"`
	OauthScopes           types.List   `tfsdk:"oauth_scopes"`
	Credentials           types.String `tfsdk:"credentials"`
	Customer              types.String `tfsdk:"customer_id"`
	ImpersonatedUserEmail types.String `tfsdk:"impersonated_user_email"`
	ServiceAccount        types.String `tfsdk:"service_account"`
}

func (p *provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"access_token": {
				Description: "A temporary [OAuth 2.0 access token] obtained from " +
					"the Google Authorization server, i.e. the `Authorization: Bearer` token used to " +
					"authenticate HTTP requests to Google Admin SDK APIs. This is an alternative to `credentials`, " +
					"and ignores the `scopes` field. If both are specified, `access_token` will be " +
					"used over the `credentials` field.",
				Type:     types.StringType,
				Optional: true,
			},

			"credentials": {
				Description: "Either the path to or the contents of a service account key file in JSON format " +
					"you can manage key files using the Cloud Console).  If not provided, the application default " +
					"credentials will be used.",
				Type:     types.StringType,
				Optional: true,
			},

			"customer_id": {
				Description: "The customer id provided with your Google Workspace subscription. It is found " +
					"in the admin console under Account Settings.",
				Type:     types.StringType,
				Optional: true,
			},

			"impersonated_user_email": {
				Description: "The impersonated user's email with access to the Admin APIs can access the Admin SDK Directory API. " +
					"`impersonated_user_email` is required for all services except group and user management.",
				Type:     types.StringType,
				Optional: true,
			},

			"oauth_scopes": {
				Description: "The list of the scopes required for your application (for a list of possible scopes, see " +
					"[Authorize requests](https://developers.google.com/admin-sdk/directory/v1/guides/authorizing))",
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional: true,
			},

			"service_account": {
				Description: "The service account used to create the provided `access_token` if authenticating using " +
					"the `access_token` method and needing to impersonate a user. This service account will require the " +
					"GCP role `Service Account Token Creator` if needing to impersonate a user.",
				Type:     types.StringType,
				Optional: true,
			},
		},
	}, nil
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	var providerConfig providerData
	diags := req.Config.Get(ctx, &providerConfig)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if providerConfig.AccessToken.Unknown {
		addCannotInterpolateInProviderBlockError(resp, "access_token")
		return
	}

	if providerConfig.Credentials.Unknown {
		addCannotInterpolateInProviderBlockError(resp, "credentials")
		return
	}

	if providerConfig.Customer.Unknown {
		addCannotInterpolateInProviderBlockError(resp, "customer_id")
		return
	}

	if providerConfig.ImpersonatedUserEmail.Unknown {
		addCannotInterpolateInProviderBlockError(resp, "impersonated_user_email")
		return
	}

	if providerConfig.OauthScopes.Unknown {
		addCannotInterpolateInProviderBlockError(resp, "oauth_scopes")
		return
	}

	if providerConfig.ServiceAccount.Unknown {
		addCannotInterpolateInProviderBlockError(resp, "service_account")
		return
	}

	// if unset, fallback to defaults
	if providerConfig.Credentials.Null {
		for _, val := range CredentialsEnvVars {
			if envVar := os.Getenv(val); envVar != "" {
				providerConfig.Credentials.Value = envVar
				break
			}
		}
	}

	if providerConfig.Customer.Null {
		providerConfig.Customer.Value = os.Getenv("GOOGLEWORKSPACE_CUSTOMER_ID")
	}

	if providerConfig.ImpersonatedUserEmail.Null {
		providerConfig.ImpersonatedUserEmail.Value = os.Getenv("GOOGLEWORKSPACE_IMPERSONATED_USER_EMAIL")
	}

	if providerConfig.OauthScopes.Null || len(providerConfig.OauthScopes.Elems) == 0 {
		providerConfig.OauthScopes = stringSliceToTypeList(DefaultClientScopes)

		providerConfig.OauthScopes.Null = false
	}

	// required if still unset
	if providerConfig.Credentials.Value == "" && providerConfig.AccessToken.Value == "" {
		addAttributeMustBeSetError(resp, "credentials or access_token")
		return
	}

	if providerConfig.Customer.Value == "" {
		addAttributeMustBeSetError(resp, "customer_id")
		return
	}

	// TODO: update this if/when the plugin-framework adds a helper function do this
	p.userAgent = fmt.Sprintf("Terraform/%s (+https://www.terraform.io) terraform-plugin-framework terraform-provider-googleworkspace/%s", req.TerraformVersion, p.version)

	// If the upstream provider SDK or HTTP client requires configuration, such
	// as authentication or logging, this is a great opportunity to do so.
	p.client = authenticateClient(ctx, providerConfig, &diags)
	p.customer = providerConfig.Customer.Value
	p.credentials = providerConfig.Credentials.Value
	p.oauthscopes = providerConfig.OauthScopes
	p.configured = true
}

// GetResources - Defines provider resources
func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		//"googleworkspace_chrome_policy":       resourceChromePolicyType{},
		//"googleworkspace_domain":              resourceDomainType{},
		//"googleworkspace_domain_alias":        resourceDomainAliasType{},
		//"googleworkspace_gmail_send_as_alias": resourceGmailSendAsAliasType{},
		//"googleworkspace_group":               resourceGroupType{},
		//"googleworkspace_group_member":        resourceGroupMemberType{},
		//"googleworkspace_group_members":       resourceGroupMembersType{},
		//"googleworkspace_group_settings":      resourceGroupSettingsType{},
		"googleworkspace_org_unit": resourceOrgUnitType{},
		//"googleworkspace_role":                resourceRoleType{},
		"googleworkspace_role_assignment": resourceRoleAssignmentType{},
		"googleworkspace_schema":          resourceSchemaType{},
		"googleworkspace_user":            resourceUserType{},
	}, nil
}

// GetDataSources - Defines provider data sources
func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		//"googleworkspace_chrome_policy_schema": dataSourceChromePolicySchemaType{},
		//"googleworkspace_domain":               dataSourceDomainType{},
		//"googleworkspace_domain_alias":         dataSourceDomainAliasType{},
		//"googleworkspace_group":                dataSourceGroupType{},
		//"googleworkspace_group_member":         dataSourceGroupMemberType{},
		//"googleworkspace_group_members":        dataSourceGroupMembersType{},
		//"googleworkspace_group_settings":       dataSourceGroupSettingsType{},
		//"googleworkspace_org_unit":             dataSourceOrgUnitType{},
		//"googleworkspace_privileges":           dataSourcePrivilegesType{},
		//"googleworkspace_role":                 dataSourceRoleType{},
		//"googleworkspace_schema":               dataSourceSchemaType{},
		//"googleworkspace_user":                 dataSourceUserType{},
		//"googleworkspace_users":                dataSourceUsersType{},
	}, nil
}

//
//func init() {
//	// Set descriptions to support markdown syntax, this will be used in document generation
//	// and the language server.
//	schema.DescriptionKind = schema.StringMarkdown
//
//	// Customize the content of descriptions when output. For example you can add defaults on
//	// to the exported descriptions if present.
//	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
//		desc := s.Description
//		if s.Default != nil {
//			desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
//		}
//		return strings.TrimSpace(desc)
//	}
//}
