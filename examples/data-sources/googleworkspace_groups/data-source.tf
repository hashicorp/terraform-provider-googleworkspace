# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "googleworkspace_groups" "my-domain-groups" {
}

output "num_groups" {
  value = length(data.googleworkspace_groups.my-domain-groups.groups)
}