# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "googleworkspace_group" "sales" {
  email = "sales@example.com"
}

output "group_name" {
  value = data.googleworkspace_group.sales.name
}