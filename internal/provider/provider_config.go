package provider

import (
	"context"
	"log"
	"net/http"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	directory "google.golang.org/api/admin/directory/v1"
)

type apiClient struct {
	client *http.Client

	ClientScopes []string
	Credentials  string
	TokenSource  oauth2.TokenSource
	UserAgent    string
}

func (c *apiClient) loadAndValidate(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(c.ClientScopes) == 0 {
		c.ClientScopes = DefaultClientScopes
	}

	if c.Credentials != "" {
		contents, _, err := pathOrContents(c.Credentials)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  err.Error(),
			})
		}

		creds, err := googleoauth.CredentialsFromJSON(ctx, []byte(contents), c.ClientScopes...)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  err.Error(),
				Detail:   contents,
			})

			return diags
		}

		c.TokenSource = creds.TokenSource

		cleanCtx := context.WithValue(ctx, oauth2.HTTPClient, cleanhttp.DefaultClient())

		// 1. OAUTH2 TRANSPORT/CLIENT - sets up proper auth headers
		client := oauth2.NewClient(cleanCtx, creds.TokenSource)

		c.client = client
	}

	return diags
}

func (c *apiClient) NewDirectoryService() *directory.Service {
	log.Printf("[INFO] Instantiating Google Admin Directory service")
	directoryService, err := directory.NewService(context.Background(), option.WithHTTPClient(c.client))
	if err != nil {
		log.Printf("[WARN] Error creating directory service: %s", err)
		return nil
	}

	return directoryService
}
