# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "googleworkspace_group" "sales" {
  email       = "sales@example.com"
  name        = "Sales"
  description = "Sales Group"

  aliases = ["paper-sales@example.com", "sales-dept@example.com"]

  timeouts {
    create = "1m"
    update = "1m"
  }
}