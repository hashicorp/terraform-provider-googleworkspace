# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "googleworkspace_users" "my-domain-users" {}

output "num_users" {
  value = length(data.googleworkspace_users.my-domain-users.users)
}