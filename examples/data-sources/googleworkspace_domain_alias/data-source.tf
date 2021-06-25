data "googleworkspace_domain_alias" "example" {
  domain_alias_name = "alias-example.com"
}

output "parent-domain" {
  value = data.googleworkspace_domain_alias.example.parent_domain_name
}