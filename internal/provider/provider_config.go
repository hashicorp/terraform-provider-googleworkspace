package googleworkspace

import (
	"context"
	"log"
	"net/http"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"

	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"

	directory "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/chromepolicy/v1"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/groupssettings/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
)

type apiClient struct {
	client *http.Client

	AccessToken           string
	ClientScopes          []string
	Credentials           string
	Customer              string
	ImpersonatedUserEmail string
	UserAgent             string
}

func (c *apiClient) loadAndValidate(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(c.ClientScopes) == 0 {
		c.ClientScopes = DefaultClientScopes
	}

	if c.AccessToken != "" {
		contents, _, err := pathOrContents(c.AccessToken)
		if err != nil {
			return diag.FromErr(err)
		}
		token := &oauth2.Token{AccessToken: contents}

		log.Printf("[INFO] Authenticating using configured Google JSON 'access_token'...")
		log.Printf("[INFO]   -- Scopes: %s", c.ClientScopes)

		if c.ImpersonatedUserEmail != "" {
			opts := []option.ClientOption{
				option.WithTokenSource(oauth2.StaticTokenSource(token)),
				option.ImpersonateCredentials(c.ImpersonatedUserEmail),
				option.WithScopes(c.ClientScopes...),
			}
			creds, err := transport.Creds(context.TODO(), opts...)
			if err != nil {
				return diag.FromErr(err)
			}
			diags = c.SetupClient(ctx, creds)
			return diags
		}

		creds := googleoauth.Credentials{
			TokenSource: oauth2.StaticTokenSource(token),
		}
		diags = c.SetupClient(ctx, &creds)
	}

	if c.Credentials != "" {
		contents, _, err := pathOrContents(c.Credentials)
		if err != nil {
			return diag.FromErr(err)
		}

		credParams := googleoauth.CredentialsParams{
			Scopes:  c.ClientScopes,
			Subject: c.ImpersonatedUserEmail,
		}

		creds, err := googleoauth.CredentialsFromJSONWithParams(ctx, []byte(contents), credParams)
		if err != nil {
			return diag.FromErr(err)
		}

		diags = c.SetupClient(ctx, creds)
	} else {
		credParams := googleoauth.CredentialsParams{
			Scopes:  c.ClientScopes,
			Subject: c.ImpersonatedUserEmail,
		}

		creds, err := googleoauth.FindDefaultCredentialsWithParams(ctx, credParams)
		if err != nil {
			return diag.FromErr(err)
		}

		diags = c.SetupClient(ctx, creds)
	}

	return diags
}

func (c *apiClient) SetupClient(ctx context.Context, creds *googleoauth.Credentials) diag.Diagnostics {
	var diags diag.Diagnostics

	cleanCtx := context.WithValue(ctx, oauth2.HTTPClient, cleanhttp.DefaultClient())

	// 1. MTLS TRANSPORT/CLIENT - sets up proper auth headers
	client, _, err := transport.NewHTTPClient(cleanCtx, option.WithTokenSource(creds.TokenSource))
	if err != nil {
		return diag.FromErr(err)
	}

	// 2. Logging Transport - ensure we log HTTP requests to admin APIs.
	loggingTransport := logging.NewTransport("Google Workspace", client.Transport)

	// Set final transport value.
	client.Transport = loggingTransport

	c.client = client
	return diags
}

func (c *apiClient) NewChromePolicyService() (*chromepolicy.Service, diag.Diagnostics) {
	var diags diag.Diagnostics

	log.Printf("[INFO] Instantiating Google Admin Chrome Policy service")

	chromePolicyService, err := chromepolicy.NewService(context.Background(), option.WithHTTPClient(c.client))
	if err != nil {
		return nil, diag.FromErr(err)
	}

	if chromePolicyService == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Directory Service could not be created.",
		})

		return nil, diags
	}

	return chromePolicyService, diags
}

func (c *apiClient) NewDirectoryService() (*directory.Service, diag.Diagnostics) {
	var diags diag.Diagnostics

	log.Printf("[INFO] Instantiating Google Admin Directory service")

	directoryService, err := directory.NewService(context.Background(), option.WithHTTPClient(c.client))
	if err != nil {
		return nil, diag.FromErr(err)
	}

	if directoryService == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Directory Service could not be created.",
		})

		return nil, diags
	}

	return directoryService, diags
}
func (c *apiClient) NewGmailService(ctx context.Context, userId string) (*gmail.Service, diag.Diagnostics) {
	var diags diag.Diagnostics

	log.Printf("[INFO] Instantiating Google Admin Gmail service")

	// the send-as-alias resource requires the oauth token impersonate the user
	// the alias is being created for.
	log.Printf("[INFO] Creating Google Admin Gmail client that impersonates %q", userId)
	newClient := &apiClient{
		Credentials:           c.Credentials,
		ClientScopes:          c.ClientScopes,
		Customer:              c.Customer,
		UserAgent:             c.UserAgent,
		ImpersonatedUserEmail: userId,
	}
	diags = newClient.loadAndValidate(ctx)
	if diags.HasError() {
		return nil, diags
	}

	gmailService, err := gmail.NewService(ctx, option.WithHTTPClient(newClient.client))
	if err != nil {
		return nil, diag.FromErr(err)
	}

	if gmailService == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Gmail Service could not be created.",
		})

		return nil, diags
	}

	return gmailService, diags
}

func (c *apiClient) NewGroupsSettingsService() (*groupssettings.Service, diag.Diagnostics) {
	var diags diag.Diagnostics

	log.Printf("[INFO] Instantiating Google Admin Groups Settings service")

	groupsSettingsService, err := groupssettings.NewService(context.Background(), option.WithHTTPClient(c.client))
	if err != nil {
		return nil, diag.FromErr(err)
	}

	if groupsSettingsService == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Groups Settings Service could not be created.",
		})

		return nil, diags
	}

	return groupsSettingsService, diags
}
