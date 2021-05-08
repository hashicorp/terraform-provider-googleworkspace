---
page_title: "Google Workspace Provider Configuration Reference"
subcategory: ""
description: |-
  Configuration reference for the Google Workspace provider for Terraform.
---

# Google Workspace Provider Configuration Reference

The `googleworkspace` provider block is used to configure the
credentials you use to authenticate with Google Workspace,
as well as the domain for your resources.

## Example Usage - Basic provider blocks

```hcl
provider "googleworkspace" {
  credentials             = "/Users/mscott/my-project-c633d7053aab.json"
  customer_id             = "A01b123xz"
  domain                  = "example.com"
  impersonated_user_email = "impersonated@example.com"
}
```

## Authentication

### Creating a Service Account and Credentials

Terraform will use a GCP service account to manage resources created by the provider. To create the resource and
generate a service account key see the documentation [here](https://developers.google.com/admin-sdk/directory/v1/guides/delegation#create_the_service_account_and_credentials).
Once the key has been generated, save the json file locally and set the `GOOGLEWORKSPACE_CREDENTIALS` environment
variable to the path of the service account key. Terraform will use that key for authentication.

### Configuring the Service Account

To access user data on a Google Workspace domain, the service account that you created needs to be granted access
by a super administrator for the domain. To delegate domain-wide authority to a service account, follow the instructions
[here](https://developers.google.com/admin-sdk/directory/v1/guides/delegation#delegate_domain-wide_authority_to_your_service_account).

* Note: The Oauth scopes granted to your service account must match, or be a superset, of the `oauth_scopes` granted to
the `googleworkspace` provider.

### Impersonating a Google Workspace User

Only users with access to the Admin APIs can access the Admin SDK Directory API, therefore your service account needs
to impersonate one of those users to access the Admin SDK Directory API. This user's email must be set in the environment
variable `GOOGLEWORKSPACE_IMPERSONATED_USER_EMAIL` or in the `impersonated_user_email` attribute in the provider.
Additionally, the user must have logged in at least once and accepted the Google Workspace Terms of Service.