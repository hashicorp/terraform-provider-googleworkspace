data "googleworkspace_user" "example" {
  primary_email = "user.with.gmail.license@example.com"
}

resource "googleworkspace_user" "alias" {
  primary_email = "alias@example.com"
  password      = "34819d7beeabb9260a5c854bc85b3e44"
  hash_function = "MD5"

  name {
    family_name = "Scott"
    given_name  = "Michael"
  }
}

resource "googleworkspace_gmail_send_as_alias" "test" {
  primary_email = data.googleworkspace_user.example.primary_email
  send_as_email = googleworkspace_user.alias.primary_email
}