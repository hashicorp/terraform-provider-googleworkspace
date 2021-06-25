resource "googleworkspace_domain_alias" "example" {
  parent_domain_name = "example.com"
  domain_alias_name  = "alias-example.com"
}