# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "googleworkspace_group_settings" "sales-settings" {
  email = "sales@example.com"
}

output "who_can_join_sales" {
  value = data.googleworkspace_group_settings.sales-settings.who_can_join
}