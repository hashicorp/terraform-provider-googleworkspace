// a test exists proving no impersonation is required for some resources
// as long as the service account has the proper role assignments

data "googleworkspace_role" "groups-admin" {
  name = "_GROUPS_ADMIN_ROLE"
}

resource "googleworkspace_role_assignment" "sa-groups-admin" {
  role_id = data.googleworkspace_role.groups-admin.id
  assigned_to = google_service_account.acctest-sa.unique_id
}
