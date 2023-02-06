# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "googleworkspace_user" "dwight" {
  primary_email = "dwight.schrute@example.com"
}

output "is_user_admin" {
  value = data.googleworkspace_user.dwight.is_admin
}