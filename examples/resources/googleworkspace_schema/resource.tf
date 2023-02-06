# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "googleworkspace_schema" "birthday" {
  schema_name = "birthday"

  fields {
    field_name = "birthday"
    field_type = "DATE"
  }
}