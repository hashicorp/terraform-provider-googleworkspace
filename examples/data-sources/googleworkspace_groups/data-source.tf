data "googleworkspace_groups" "my-domain-groups" {
}

output "num_groups" {
  value = length(data.googleworkspace_groups.my-domain-groups.groups)
}