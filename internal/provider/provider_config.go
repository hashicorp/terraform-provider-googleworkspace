package googleworkspace

import (
	"context"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"

	directory "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/chromepolicy/v1"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/groupssettings/v1"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
)

func authenticateClient(ctx context.Context, providerConfig providerData, diags *diag.Diagnostics) *http.Client {
	oauthScopes := make([]string, len(providerConfig.OauthScopes.Elems))
	diags.Append(providerConfig.OauthScopes.ElementsAs(ctx, &oauthScopes, false)...)
	if diags.HasError() {
		return nil
	}

	if providerConfig.AccessToken.Value != "" {
		contents, _, err := pathOrContents(providerConfig.AccessToken.Value)
		if err != nil {
			diags.AddError(
				"Unexpected error reading provider `access_token`",
				err.Error(),
			)
			return nil
		}
		token := &oauth2.Token{AccessToken: contents}

		log.Printf("[INFO] Authenticating using configured Google JSON 'access_token'...")
		log.Printf("[INFO]   -- Scopes: %+v", oauthScopes)

		if providerConfig.ImpersonatedUserEmail.Value != "" {
			if providerConfig.ServiceAccount.Value == "" {
				diags.AddError(
					"Invalid provider config",
					"`service_account` is required to impersonate a user with the `access_token` authentication",
				)

				return nil
			}

			tokenSource, err := impersonate.CredentialsTokenSource(context.TODO(), impersonate.CredentialsConfig{
				TargetPrincipal: providerConfig.ServiceAccount.Value,
				Scopes:          oauthScopes,
				Subject:         providerConfig.ImpersonatedUserEmail.Value,
			}, option.WithTokenSource(oauth2.StaticTokenSource(token)))
			if err != nil {
				diags.AddError(
					"Unexpected error creating credentials token source with `access_token`",
					err.Error(),
				)
			}

			creds := googleoauth.Credentials{
				TokenSource: tokenSource,
			}
			return Client(ctx, &creds, diags)
		}

		creds := googleoauth.Credentials{
			TokenSource: oauth2.StaticTokenSource(token),
		}
		return Client(ctx, &creds, diags)
	}

	if providerConfig.Credentials.Value != "" {
		contents, _, err := pathOrContents(providerConfig.Credentials.Value)
		if err != nil {
			diags.AddError(
				"Unexpected error reading provider `credentials`",
				err.Error(),
			)
			return nil
		}

		credParams := googleoauth.CredentialsParams{
			Scopes:  oauthScopes,
			Subject: providerConfig.ImpersonatedUserEmail.Value,
		}

		creds, err := googleoauth.CredentialsFromJSONWithParams(ctx, []byte(contents), credParams)
		if err != nil {
			diags.AddError(
				"Unexpected error creating credentials with provider `credentials`",
				err.Error(),
			)
			return nil
		}
		if creds == nil {
			diags.AddError("creds are nil", "object returned is nil")
			return nil
		}

		return Client(ctx, creds, diags)
	} else {
		credParams := googleoauth.CredentialsParams{
			Scopes:  oauthScopes,
			Subject: providerConfig.ImpersonatedUserEmail.Value,
		}

		creds, err := googleoauth.FindDefaultCredentialsWithParams(ctx, credParams)
		if err != nil {
			diags.AddError(
				"Unexpected error finding default credentials",
				err.Error(),
			)
			return nil
		}

		return Client(ctx, creds, diags)
	}

}

func Client(ctx context.Context, creds *googleoauth.Credentials, diags *diag.Diagnostics) *http.Client {
	cleanCtx := context.WithValue(ctx, oauth2.HTTPClient, cleanhttp.DefaultClient())

	// 1. MTLS TRANSPORT/CLIENT - sets up proper auth headers
	client, _, err := transport.NewHTTPClient(cleanCtx, option.WithTokenSource(creds.TokenSource))
	if err != nil {
		diags.AddError(
			"Unexpected error creating new HTTP client",
			err.Error(),
		)
		return nil
	}

	// 2. Logging Transport - ensure we log HTTP requests to admin APIs.
	scrubbedLoggingTransport := NewTransportWithScrubbedLogs("Google Workspace", client.Transport)

	// 3. Retry Transport - retries common temporary errors
	// Keep order for wrapping logging so we log each retried request as well.
	// This value should be used if needed to create shallow copies with additional retry predicates.
	// See ClientWithAdditionalRetries
	retryTransport := NewTransportWithDefaultRetries(scrubbedLoggingTransport)

	// Set final transport value.
	client.Transport = retryTransport

	return client
}

func (p *provider) NewChromePolicyService(diags *diag.Diagnostics) *chromepolicy.Service {
	log.Printf("[INFO] Instantiating Google Admin Chrome Policy service")
	chromePolicyService, err := chromepolicy.NewService(context.Background(), option.WithHTTPClient(p.client))
	if err != nil {
		diags.AddError("could not instantiate Chrome Policy service", err.Error())
	}

	if chromePolicyService == nil {
		diags.AddError("Chrome Policy service was not created", "Chrome Policy service is nil")
	}

	return chromePolicyService
}

func (p *provider) NewDirectoryService(diags *diag.Diagnostics) *directory.Service {
	log.Printf("[INFO] Instantiating Google Admin Directory service")
	directoryService, err := directory.NewService(context.Background(), option.WithHTTPClient(p.client))
	if err != nil {
		diags.AddError("could not instantiate Admin Directory service", err.Error())
	}

	if directoryService == nil {
		diags.AddError("Admin Directory service was not created", "Admin Directory service is nil")
	}

	return directoryService
}

func (p *provider) NewGmailService(diags *diag.Diagnostics) *gmail.Service {
	log.Printf("[INFO] Instantiating Google Admin Gmail service")

	gmailService, err := gmail.NewService(context.Background(), option.WithHTTPClient(p.client))
	if err != nil {
		diags.AddError("could not instantiate Admin Gmail service", err.Error())
		return nil
	}

	if gmailService == nil {
		diags.AddError("Admin Gmail service was not created", "Admin Gmail service is nil")
		return nil
	}

	return gmailService
}

func (p *provider) NewGroupsSettingsService(diags *diag.Diagnostics) *groupssettings.Service {
	log.Printf("[INFO] Instantiating Google Admin Groups Settings service")
	groupsSettingsService, err := groupssettings.NewService(context.Background(), option.WithHTTPClient(p.client))
	if err != nil {
		diags.AddError("could not instantiate Admin Groups Settings service", err.Error())
	}

	if groupsSettingsService == nil {
		diags.AddError("Admin Groups Settings service was not created", "Admin Groups Settings service is nil")
	}

	return groupsSettingsService
}
