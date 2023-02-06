# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "googleworkspace_group" "sales" {
  email = "sales@example.com"
}

resource "googleworkspace_user" "michael" {
  primary_email = "michael.scott@example.com"
  password      = "34819d7beeabb9260a5c854bc85b3e44"
  hash_function = "MD5"

  name {
    family_name = "Scott"
    given_name  = "Michael"
  }
}

resource "googleworkspace_user" "frank" {
  primary_email = "frank.scott@example.com"
  password      = "2095312189753de6ad47dfe20cbe97ec"
  hash_function = "MD5"

  name {
    family_name = "Scott"
    given_name  = "Frank"
  }
}

resource "googleworkspace_group_members" "sales" {
  group_id = googleworkspace_group.sales.id

  members {
    email = googleworkspace_user.michael.primary_email
    role  = "MANAGER"
  }

  members {
    email = googleworkspace_user.frank.primary_email
    role  = "MEMBER"
  }
}
