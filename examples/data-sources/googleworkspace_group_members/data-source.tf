# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "googleworkspace_group" "sales" {
  email = "sales@example.com"
}

data "googleworkspace_group_members" "sales" {
  group_id = data.googleworkspace_group.sales.id
}

output "group_members" {
  value = data.googleworkspace_group_members.sales.members
}
