## 0.5.1 (Unreleased)

BUG FIXES:

* directory: fixed eventual consistency for `googleworkspace_group_member` [GH-186]
* directory: fixed bug where `googleworkspace_group_members` would error if a member already existed [GH-194]
* directory: fixed bug where `googleworkspace_group_members` would error if a `members` was empty [GH-193]

## 0.5.0 (October 13, 2021)

FEATURES:

* **New Resource:**   `googleworkspace_group_members` ([#155](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/155))

* **New Datasource:** `googleworkspace_group_members` ([#155](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/155))

IMPROVEMENTS:

* provider: added ability to authenticate using user credentials ([#156](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/156))
* provider: added ability to authenticate using `access_token` rather than just credentials ([#165](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/165))
* provider: added retryTransport that will retry after common Google errors (like network errors, rate limiting errors, etc) ([#163](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/163))

BUG FIXES:

* directory: fixed bug where `googleworkspace_group_member` would not force new change on change of email ([#152](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/152))
* directory: fixed bug where `googleworkspace_user` would show permadiff on user alias emails from secondary domains ([#154](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/154))

## 0.4.1 (August 16, 2021)

BUG FIXES:

* provider: fixed validation of credentials ([#126](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/126))
* directory: fixed bug where `googleworkspace_user.password` was always required, which would break on import, now it is only required on create. ([#125](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/125))
* directory: fixed permadiff on`googleworkspace_user.custom_schemas` ([#129](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/129))
* directory: fixed bug where different fields on `googleworkspace_user` would error if the value was empty. ([#133](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/133))
* directory: fixed bugs on `googleworkspace_group`, `googleworkspace_group_member` and `googleworkspace_org_unit` where changes made out of band would not refresh appropriately. ([#136](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/136))
* directory: changed nested `type` fields on `googleworkspace_org_user` from optional to required. ([#139](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/139))
* gmail: fixed bugs on `googleworkspace_gmail_send_as_alias` where changes made out of band would not refresh appropriately. ([#136](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/136))

## 0.4.0 (July 27, 2021)

FEATURES:

* **New Resource:** `googleworkspace_gmail_send_as_alias` ([#122](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/122))

IMPROVEMENTS:

* provider: allow impersonation to be unset when service account has sufficient role assignments. ([#115](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/115))

BUG FIXES:

* groups settings: fixed bug where consistency check was prone to failure/timeout. ([#119](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/119))
* directory: fixed bug where `googleworkspace_schema.fields.indexed` would break if it was nil. ([#108](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/108))
* chrome: fixed validation bug on message type schema values, ([#116](https://github.com/hashicorp/terraform-provider-googleworkspace/issues/116))

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
