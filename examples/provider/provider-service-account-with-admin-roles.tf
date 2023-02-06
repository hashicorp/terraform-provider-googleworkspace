# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Auth method: Admin roles applied directly to a service account
provider "googleworkspace" {
  credentials = "/Users/mscott/my-project-c633d7053aab.json"
  customer_id = "A01b123xz"
  oauth_scopes = [
    "https://www.googleapis.com/auth/admin.directory.user",
    "https://www.googleapis.com/auth/admin.directory.userschema",
    # include scopes as needed
  ]
}