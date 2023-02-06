# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "googleworkspace_chrome_policy_schema" "example" {
  schema_name = "chrome.printers.AllowForUsers"
}

output "field_descriptions" {
  value = data.googleworkspace_chrome_policy_schema.example.field_descriptions
}