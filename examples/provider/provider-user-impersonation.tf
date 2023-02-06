# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Auth method: Domain-wide delegation and user impersonation
provider "googleworkspace" {
  credentials             = "/Users/mscott/my-project-c633d7053aab.json"
  customer_id             = "A01b123xz"
  impersonated_user_email = "impersonated@example.com"
  oauth_scopes = [
    "https://www.googleapis.com/auth/admin.directory.user",
    "https://www.googleapis.com/auth/admin.directory.userschema",
    # include scopes as needed
  ]
}