---
layout: ""
page_title: "Google Workspace Provider"
subcategory: ""
description: |-
Terraform Google Workspace Provider Upgrade Guide
---

# Terraform Google Workspace Provider Upgrade Guide

The Google Workspace provider for Terraform is the first official provider to maintain
resources within Google Workspace. Prior to the official provider was a [community-maintained
GSuite provider](https://github.com/DeviaVir/terraform-provider-gsuite). While we kept feature-parity with the latest
version of the GSuite provider, there exist some changes that you will need to consider when switching. This guide
is intended to help with that process.

## Why change?

We introduced the Google Workspace provider in order to provide support for Google Workspace
resources at an official capacity. With this release, the community-maintained GSuite provider
chose to deprecate itself and refer users to this provider.

In addition to feature-parity, we added some new data sources and added some attributes to 
some of the already existing resources. While many resource configurations stayed the same 
between the two providers, you will see mostly small changes, with one major difference in 
regard to user schemas.


## Upgrade Topics

<!-- TOC depthFrom:2 depthTo:2 -->

- [Provider Configuration](#provider-configuration)
- [Data Sources](#data-sources)
- [Data_Source: `gsuite_user_attributes`](#data-source-gsuite_user_attributes)
- [Resource: `gsuite_user_attributes`](#resource-gsuite_user_attributes)
- [Resource: `gsuite_group_members`](#resource-gsuite_group_members)
- [Resource: `googleworkspace_group_member`](#resource-googleworkspace_group_member)
- [Resource: `googleworkspace_group_settings`](#resource-googleworkspace_group_settings)
- [Resource: `googleworkspace_user`](#resource-googleworkspace_user)
- [Resource: `googleworkspace_schema`](#resource-googleworkspace_schema)
- [User Attributes](#user-attributes)

<!-- /TOC -->

## Provider Configuration

While the required attributes for defining the provider have not changed,
we have changed the environment variable names and removed the optional 
`timeout_minutes` and `update_existing`.

The names of the environment variables changed slightly. See the table below 
for any changes you would need to make to your environment variables

~> Note: `GOOGLE_CREDENTIALS` was added as an option in version `0.2.0` of the Google Workspace provider.

|                         | GSuite Provider                                                                                               | Google Workspace Provider                                                                                    |
|-------------------------|---------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------|
| Credentials             | `GOOGLE_CREDENTIALS`,  `GOOGLE_CLOUD_KEYFILE_JSON`,  `GOOGLE_KEYFILE_JSON`,  `GOOGLE_APPLICATION_CREDENTIALS` | `GOOGLEWORKSPACE_CREDENTIALS`, `GOOGLEWORKSPACE_CLOUD_KEYFILE_JSON`, `GOOGLE_CREDENTIALS` (>= 0.2.0) |
| Impersonated User Email | `IMPERSONATED_USER_EMAIL`                                                                                     | `GOOGLEWORKSPACE_IMPERSONATED_USER_EMAIL`                                                        |
| Customer Id             | N/A                                                                                                           | `GOOGLEWORKSPACE_CUSTOMER_ID`                                                                    |

While we removed `timeout_minutes`, the Google Workspace provider now allows a timeout block 
to be set on each particular resource. For example, given this previous configuration:

```hcl
provider "gsuite" {
  # ... other configuration ...

  timeout_minutes = 1
}
```

An updated configuration:

```hcl
resource "googleworkspace_group" "group" {
  # ... other configuration ...

  timeouts {
    create = "5m"
    update = "5m"
  }
}
```

Although the option for `update_existing` was not included in the official provider, 
the functionality is being tracked in an [open issue for Terraform](https://github.com/hashicorp/terraform/issues/19017).

## Data Sources

See the `Resource` sections in this document for properties that may have been
added, removed or changed.

## Data_Source: `gsuite_user_attributes`

## Resource: `gsuite_user_attributes`

We removed the above data source and resource in favor of a different approach to defining user custom schemas. 
See how `googleworkspace_user.custom_schemas` are defined [below](#user-attributes).

## Resource: `gsuite_group_members`

We removed this resource in favor of defining multiple `googleworkspace_group_member` resources. 
For example, given this previous configuration:

```hcl
resource "gsuite_group_members" "testing_team_members" {
  group_email = gsuite_group.testing_team.email

  member {
    email = "a@xxx.com"
    role  = "MEMBER"
  }

  member {
    email = "b@xxx.com"
    role  = "OWNER"
  }
}
```

An updated configuration could look like this:

```hcl
locals {
  members = {
    "a@xxx.com" = "MEMBER"
    "b@xxx.com" = "OWNER"
  }
}

resource "googleworkspace_group_member" "testing_team_members" {
  for_each = local.members
  group_id = googleworkspace_group.testing_team.id

  email = each.key
  role  = each.value
}
```

## Resource: `googleworkspace_group_member`

There are a few minor changes to `googleworkspace_group_member`. The attribute `group` changed from accepting
a group's `email` to a group's `group_id`. Also the read-only field, `kind`, is no longer available.

The read-only field `id` is the string that will be used in order to import a group member and is of the format 
`groups/<group_id>/members/<member_id>` where `member_id` (in this case) is the id of the customer, group or user.

## Resource: `googleworkspace_group_settings`

The biggest change to this resource is that `description` of the group cannot be changed on `googleworkspace_group_settings`.
Changes to a group's `description` must be done on the `googleworkspace_group` resource.

Another notable change here is that `is_archived` can be set to `true` or `false` rather than being read-only.

While `kind` is no longer a read-only property, `custom_roles_enabled_for_settings_to_be_merged` is available.

## Resource: `googleworkspace_user`

While the majority of changes here include additional attributes, there are some name changes. Specifically, `is_ip_whitelisted` 
changed to `ip_allowlist`, `is_suspended` to `suspended`, `2s_enforced` to `is_enforced_in_2_step_verification` and 
`2s_enrolled` to `is_enrolled_in_2_step_verification`. 

The biggest change, however, is how `googleworkspace_user.custom_schemas` are now defined. See more information 
[below](#user-attributes).

The read-only field `id` is the string that will be used in order to import a user.

## Resource: `googleworkspace_schema`

The obvious change here is the name from `gsuite_user_schema` to `googleworkspace_schema`, but beyond that some minor 
changes where the `field` attribute is now `fields` and `field.range` is now `fields.numeric_indexing_spec` to better 
align with the API names.

## User Attributes

Perhaps the biggest change of all is to the provider defines custom schemas. In the GSuite provider, one would define a 
user schema (`gsuite_user_schema`), then define a user attributes data source (`gsuite_user_attributes`) that would give 
values to the schema defined by `gsuite_user_schema`, then combine the two resources into a new resource, 
`gsuite_user_attributes`, along with a user's primary email address.

In the official Google Workspace provider, after a schema is defined (`googleworkspace_schema`), all attributes are defined 
on the user itself (`googleworkspace_user`). In the `custom_schemas` block, one will pass in a `schema_name` which will 
reference the name of the previously defined `googleworkspace_schema`, and `schema_values`, which is a map of the schema 
fields' names to their jsonencoded values.

An example may be more helpful. Given this previous configuration:

```hcl
resource "gsuite_user_schema" "details" {
  schema_name  = "additional-details"

  field {
    field_type       = "PHONE"
    field_name       = "internal-phone"
    display_name     = "Internal Phone"
    indexed          = true
    read_access_type = "ALL_DOMAIN_USERS"
    multi_valued     = false
  }
}

data "gsuite_user_attributes" "details" {
  phone {
    name  = "internal-phone"
    value = "555-555-5555"
  }
}

resource "gsuite_user" "user" {
  primary_email = "flast@domain.ext"

  name {
    given_name  = "First"
    family_name = "Last"
  }
}

resource "gsuite_user_attributes" "user_attributes" {
  primary_email = gsuite_user.user.primary_email
  custom_schema {
    name  = gsuite_user_schema.details.schema_name
    value = data.gsuite_user_attributes.details.json
  }
}
```

An updated configuration would look like this:

```hcl
resource "googleworkspace_schema" "details" {
  schema_name  = "additional-details"

  fields {
    field_type       = "PHONE"
    field_name       = "internal-phone"
    display_name     = "Internal Phone"
    indexed          = true
    read_access_type = "ALL_DOMAIN_USERS"
    multi_valued     = false
  }
}

resource "googleworkspace_user" "user" {
  primary_email = "flast@domain.ext"

  name {
    given_name  = "First"
    family_name = "Last"
  }
  
  custom_schemas {
    schema_name = googleworkspace_schema.details.schema_name
    schema_values = {
      "internal-phone" = jsonencode("555-555-555")
    }
  }
}
```