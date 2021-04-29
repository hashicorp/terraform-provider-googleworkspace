data "googleworkspace_user" "michael" {
  primary_email = "michael.scott@example.com"
}

output "is_user_admin" {
  value = data.googleworkspace_user.michael.is_admin
}