resource "googleworkspace_user" "dwight" {
  primary_email = "dwight.schrute@example.com"
  password      = "34819d7beeabb9260a5c854bc85b3e44"
  hash_function = "MD5"

  name {
    family_name = "Schrute"
    given_name  = "Dwight"
  }
}

data "googleworkspace_role" "groups-admin" {
  name = "_GROUPS_ADMIN_ROLE"
}

resource "googleworkspace_role_assignment" "dwight-ra" {
  role_id     = data.googleworkspace_role.groups-admin.id
  assigned_to = googleworkspace_user.dwight.id
}