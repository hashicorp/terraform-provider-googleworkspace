# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "googleworkspace_org_unit" "org" {
  name                 = "dunder-mifflin"
  description          = "paper company"
  parent_org_unit_path = "/"
}