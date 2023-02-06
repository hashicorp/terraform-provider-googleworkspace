# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "googleworkspace_domain_alias" "example" {
  parent_domain_name = "example.com"
  domain_alias_name  = "alias-example.com"
}