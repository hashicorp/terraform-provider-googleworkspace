## 0.4.0 (Unreleased)

IMPROVEMENTS:

* provider: allow impersonation to be unset when service account has sufficient role assignments. [GH-115]

BUG FIXES:

* directory: fixed bug where `googleworkspace_schema.fields.indexed` would break if it was nil. [GH-108]
* chrome: fixed validation bug on message type schema values, [GH-116]

## 0.3.0 (July 07, 2021)

FEATURES:

* **New Resource:** `googleworkspace_chrome_policy` ([#97](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/97))
* **New Resource:** `googleworkspace_domain_alias` ([#92](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/92))

* **New Datasource:**   `googleworkspace_chrome_policy_schema` ([#97](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/97))
* **New Datasource:**   `googleworkspace_domain_alias` ([#92](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/92))

IMPROVEMENTS:

* provider: added support for using the `GOOGLE_CREDENTIALS` environment variable for authentication ([#87](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/87))

BUG FIXES:

* all: added logging of the http requests/responses ([#93](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/93))

## 0.2.0 (June 21, 2021)

FEATURES:

* **New Resource:** `googleworkspace_org_unit` ([#63](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/63))
* **New Resource:** `googleworkspace_role` ([#66](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/66))
* **New Resource:** `googleworkspace_role_assignment` ([#66](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/66))

* **New Datasource:**   `googleworkspace_org_unit` ([#63](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/63))
* **New Datasource:**   `googleworkspace_role` ([#66](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/66))
* **New Datasource:**   `googleworkspace_privileges` ([#82](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/82))

## 0.1.0 (June 03, 2021)

FEATURES:

* **New Resource:** `googleworkspace_domain` ([#12](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/12))
* **New Resource:** `googleworkspace_group` ([#18](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/18))
* **New Resource:** `googleworkspace_group_member` ([#31](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/31))
* **New Resource:** `googleworkspace_group_settings` ([#29](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/29))
* **New Resource:** `googleworkspace_schema` ([#20](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/20))
* **New Resource:** `googleworkspace_user` ([#15](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/15))

* **New Datasource:**   `googleworkspace_domain` ([#12](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/12))
* **New Datasource:**   `googleworkspace_group` ([#18](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/18))
* **New Datasource:**   `googleworkspace_group_member` ([#31](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/31))
* **New Datasource:**   `googleworkspace_group_settings` ([#32](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/32))
* **New Datasource:**   `googleworkspace_schema` ([#20](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/20))
* **New Datasource:**   `googleworkspace_user` ([#15](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/15))
