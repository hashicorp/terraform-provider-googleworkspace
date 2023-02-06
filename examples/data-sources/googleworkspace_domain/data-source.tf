# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "googleworkspace_domain" "example" {
  domain_name = "example.com"
}

output "domain_verified" {
  value = data.googleworkspace_domain.example.verified
}