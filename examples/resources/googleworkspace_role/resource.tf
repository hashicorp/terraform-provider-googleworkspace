resource "googleworkspace_role" "enterprise-app-manager" {
  name = "enterprise_app_manager"

  privileges {
    service_id     = "02w5ecyt3pkeyqi"
    privilege_name = "MANAGE_ENTERPRISE_PRIVATE_APPS"
  }
}