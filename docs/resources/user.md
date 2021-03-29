---
page_title: "googleworkspace_user Resource - terraform-provider-googleworkspace"
subcategory: ""
description: |-
  User resource manages Google Workspace Users.
---

# Resource `googleworkspace_user`

User resource manages Google Workspace Users.



## Schema

### Required

- **name** (Block List, Min: 1, Max: 1) Holds the given and family names of the user, and the read-only fullName value.The maximum number of characters in the givenName and in the familyName values is 60.In addition, name values support unicode/UTF-8 characters, and can contain spaces, letters (a-z),numbers (0-9), dashes (-), forward slashes (/), and periods (.).Maximum allowed data size for this field is 1Kb. (see [below for nested schema](#nestedblock--name))
- **password** (String, Sensitive) Stores the password for the user account. A password can contain any combination ofASCII characters. A minimum of 8 characters is required. The maximum length is 100 characters.
- **primary_email** (String) The user's primary email address. The primaryEmail must be unique and cannot be an aliasof another user.

### Optional

- **addresses** (Block List) A list of the user's addresses. The maximum allowed data size is 10Kb. (see [below for nested schema](#nestedblock--addresses))
- **archived** (Boolean) Indicates if user is archived.
- **change_password_at_next_login** (Boolean) Indicates if the user is forced to change their password at next login. This settingdoesn't apply when the user signs in via a third-party identity provider.
- **emails** (Block List) A list of the user's email addresses. The maximum allowed data size is 10Kb. (see [below for nested schema](#nestedblock--emails))
- **external_ids** (Block List) A list of external IDs for the user, such as an employee or network ID. The maximum allowed data size is 2Kb. (see [below for nested schema](#nestedblock--external_ids))
- **hash_function** (String) Stores the hash format of the password property. We recommend sending the passwordproperty value as a base 16 bit hexadecimal-encoded hash value. Set the hashFunction valuesas either the SHA-1, MD5, or crypt hash format.
- **ims** (Block List) The user's Instant Messenger (IM) accounts. A user account can have multiple imsproperties. But, only one of these ims properties can be the primary IM contact.The maximum allowed data size is 2Kb. (see [below for nested schema](#nestedblock--ims))
- **ip_whitelisted** (Boolean) If true, the user's IP address is whitelisted.
- **is_admin** (Boolean) Indicates a user with super admininistrator privileges.
- **keywords** (Block List) A list of the user's keywords. The maximum allowed data size is 1Kb. (see [below for nested schema](#nestedblock--keywords))
- **languages** (Block List) A list of the user's languages. The maximum allowed data size is 1Kb. (see [below for nested schema](#nestedblock--languages))
- **locations** (Block List) A list of the user's locations. The maximum allowed data size is 10Kb. (see [below for nested schema](#nestedblock--locations))
- **org_unit_path** (String) The full path of the parent organization associated with the user.If the parent organization is the top-level, it is represented as a forward slash (/).
- **organizations** (Block List) A list of organizations the user belongs to. The maximum allowed data size is 10Kb. (see [below for nested schema](#nestedblock--organizations))
- **phones** (Block List) Holds the given and family names of the user, and the read-only fullName value.The maximum number of characters in the givenName and in the familyName values is 60.In addition, name values support unicode/UTF-8 characters, and can contain spaces, letters (a-z),numbers (0-9), dashes (-), forward slashes (/), and periods (.).Maximum allowed data size for this field is 1Kb. (see [below for nested schema](#nestedblock--phones))
- **posix_accounts** (Block List) A list of POSIX account information for the user. (see [below for nested schema](#nestedblock--posix_accounts))
- **recovery_email** (String) Recovery email of the user.
- **recovery_phone** (String) Recovery phone of the user. The phone number must be in the E.164 format,starting with the plus sign (+). Example: +16506661212.
- **relations** (Block List) A list of the user's relationships to other users. The maximum allowed data size for this field is 2Kb. (see [below for nested schema](#nestedblock--relations))
- **ssh_public_keys** (Block List) A list of SSH public keys. The maximum allowed data size is 10Kb. (see [below for nested schema](#nestedblock--ssh_public_keys))
- **suspended** (Boolean) Indicates if user is suspended.
- **websites** (Block List) A list of the user's websites. The maximum allowed data size is 2Kb. (see [below for nested schema](#nestedblock--websites))

### Read-only

- **agreed_to_terms** (Boolean) This property is true if the user has completed an initial login and accepted theTerms of Service agreement.
- **aliases** (List of String) asps.list of the user's alias email addresses.
- **creation_time** (String) The time the user's account was created. The value is in ISO 8601 date and time format.The time is the complete date plus hours, minutes, and seconds in the formYYYY-MM-DDThh:mm:ssTZD. For example, 2010-04-05T17:30:04+01:00.
- **customer_id** (String) The customer ID to retrieve all account users. You can use the alias my_customer torepresent your account's customerId. As a reseller administrator, you can use the resoldcustomer account's customerId. To get a customerId, use the account's primary domain in thedomain parameter of a users.list request.
- **deletion_time** (String) The time the user's account was deleted. The value is in ISO 8601 date and time format.The time is the complete date plus hours, minutes, and seconds in the form YYYY-MM-DDThh:mm:ssTZD.For example 2010-04-05T17:30:04+01:00.
- **etag** (String) ETag of the resource.
- **id** (String) The unique ID for the user.
- **is_delegated_admin** (Boolean) Indicates if the user is a delegated administrator.
- **is_enforced_in_2_step_verification** (Boolean) Is 2-step verification enforced.
- **is_enrolled_in_2_step_verification** (Boolean) Is enrolled in 2-step verification.
- **is_mailbox_setup** (Boolean) Indicates if the user's Google mailbox is created. This property is only applicableif the user has been assigned a Gmail license.
- **last_login_time** (String) The last time the user logged into the user's account. The value is in ISO 8601 dateand time format. The time is the complete date plus hours, minutes, and secondsin the form YYYY-MM-DDThh:mm:ssTZD. For example, 2010-04-05T17:30:04+01:00.
- **non_editable_aliases** (List of String) asps.list of the user's non-editable alias email addresses. These are typically outsidethe account's primary domain or sub-domain.
- **suspension_reason** (String) Has the reason a user account is suspended either by the administrator or by Google atthe time of suspension. The property is returned only if the suspended property is true.
- **thumbnail_photo_etag** (String) ETag of the user's photo
- **thumbnail_photo_url** (String) Photo Url of the user.

<a id="nestedblock--name"></a>
### Nested Schema for `name`

Required:

- **family_name** (String) The user's last name.

Optional:

- **given_name** (String) The user's first name.

Read-only:

- **full_name** (String) The user's full name formed by concatenating the first and last name values.


<a id="nestedblock--addresses"></a>
### Nested Schema for `addresses`

Optional:

- **country** (String) Country
- **country_code** (String) The country code. Uses the ISO 3166-1 standard.
- **custom_type** (String) If the address type is custom, this property contains the custom value.
- **extended_address** (String) For extended addresses, such as an address that includes a sub-region.
- **formatted** (String) A full and unstructured postal address. This is not synced with thestructured address fields.
- **locality** (String) The town or city of the address.
- **po_box** (String) The post office box, if present.
- **postal_code** (String) The ZIP or postal code, if applicable.
- **primary** (Boolean) If this is the user's primary address. The addresses list may containonly one primary address.
- **region** (String) The abbreviated province or state.
- **source_is_structured** (Boolean) Indicates if the user-supplied address was formatted.Formatted addresses are not currently supported.
- **street_address** (String) The street address, such as 1600 Amphitheatre Parkway.Whitespace within the string is ignored; however, newlines are significant.
- **type** (String) The address type.Acceptable values: `custom`, `home`, `other`, `work`.


<a id="nestedblock--emails"></a>
### Nested Schema for `emails`

Optional:

- **address** (String) The user's email address. Also serves as the email ID.This value can be the user's primary email address or an alias.
- **custom_type** (String) If the value of type is custom, this property containsthe custom type string.
- **primary** (Boolean) Indicates if this is the user's primary email.Only one entry can be marked as primary.
- **type** (String) The type of the email account.Acceptable values: `custom`, `home`, `other`, `work`.


<a id="nestedblock--external_ids"></a>
### Nested Schema for `external_ids`

Required:

- **value** (String) The value of the ID.

Optional:

- **custom_type** (String) If the value of type is custom, this property containsthe custom type string.
- **type** (String) The type of the email account.Acceptable values: `custom`, `home`, `other`, `work`.


<a id="nestedblock--ims"></a>
### Nested Schema for `ims`

Optional:

- **custom_protocol** (String) If the protocol value is custom_protocol, this property holds the customprotocol's string.
- **custom_type** (String) If the IM type is custom, this property holds the custom type string.
- **im** (String) The user's IM network ID.
- **primary** (Boolean) If this is the user's primary IM.Only one entry in the IM list can have a value of true.
- **protocol** (String) An IM protocol identifies the IM network.The value can be a custom network or the standard network.Acceptable values: `aim`, `custom_protocol`, `gtalk`, `icq`, `jabber`,`msn`, `net_meeting`, `qq`, `skype`, `yahoo`.
- **type** (String) Acceptable values: `home`, `callback`, `other`, `work`.


<a id="nestedblock--keywords"></a>
### Nested Schema for `keywords`

Required:

- **value** (String) Keyword.

Optional:

- **custom_type** (String) Custom Type.
- **type** (String) Each entry can have a type which indicates standard type of that entry.For example, keyword could be of type occupation or outlook. In addition to thestandard type, an entry can have a custom type and can give it any name. Such typesshould have the CUSTOM value as type and also have a customType value.Acceptable values: `custom`, `mission`, `occupation`, `outlook`


<a id="nestedblock--languages"></a>
### Nested Schema for `languages`

Optional:

- **custom_language** (String) Other language. A user can provide their own language name if there is nocorresponding Google III language code. If this is set, LanguageCode can't be set.
- **language_code** (String) Language Code. Should be used for storing Google III LanguageCode stringrepresentation for language. Illegal values cause SchemaException.


<a id="nestedblock--locations"></a>
### Nested Schema for `locations`

Optional:

- **area** (String) Textual location. This is most useful for display purposes to conciselydescribe the location. For example, Mountain View, CA or Near Seattle.
- **building_id** (String) Building identifier.
- **custom_type** (String) If the location type is custom, this property contains the custom value.
- **desk_code** (String) Most specific textual code of individual desk location.
- **floor_name** (String) Floor name/number.
- **floor_section** (String) Floor section. More specific location within the floor. For example,if a floor is divided into sections A, B, and C, this field would identify oneof those values.
- **type** (String) The location type.Acceptable values: `custom`, `default`, `desk`


<a id="nestedblock--organizations"></a>
### Nested Schema for `organizations`

Optional:

- **cost_center** (String) The cost center of the user's organization.
- **custom_type** (String) If the value of type is custom, this property contains the custom value.
- **department** (String) Specifies the department within the organization, such as sales or engineering.
- **description** (String) The description of the organization.
- **domain** (String) The domain the organization belongs to.
- **full_time_equivalent** (Number) The full-time equivalent millipercent within the organization (100000 = 100%)
- **location** (String) The physical location of the organization.This does not need to be a fully qualified address.
- **name** (String) The name of the organization.
- **primary** (Boolean) Indicates if this is the user's primary organization. A user may only have one primary organization.
- **symbol** (String) Text string symbol of the organization. For example,the text symbol for Google is GOOG.
- **title** (String) The user's title within the organization. For example, member or engineer.
- **type** (String) The type of organization.Acceptable values: `domain_only`, `school`, `unknown`, `work`.


<a id="nestedblock--phones"></a>
### Nested Schema for `phones`

Required:

- **value** (String) A human-readable phone number. It may be in any telephone number format.

Optional:

- **custom_type** (String) If the value of type is custom, this property contains the custom type.
- **primary** (Boolean) Indicates if this is the user's primary phone number.A user may only have one primary phone number.
- **type** (String) The type of phone number.Acceptable values: `assistant`, `callback`, `car`, `company_main`, `custom`, `grand_central`, `home`, `home_fax`, `isdn`, `main`, `mobile`, `other`,`other_fax`, `pager`, `radio`, `telex`, `tty_tdd`, `work`, `work_fax`, `work_mobile`,`work_pager`.


<a id="nestedblock--posix_accounts"></a>
### Nested Schema for `posix_accounts`

Optional:

- **account_id** (String) A POSIX account field identifier.
- **gecos** (String) The GECOS (user information) for this account.
- **gid** (String) The default group ID.
- **home_directory** (String) The path to the home directory for this account.
- **operating_system_type** (String) The operating system type for this account.Acceptable values: `linux`, `unspecified`, `windows`.
- **primary** (Boolean) If this is user's primary account within the SystemId.
- **shell** (String) The path to the login shell for this account.
- **system_id** (String) System identifier for which account Username or Uid apply to.
- **uid** (String) The POSIX compliant user ID.
- **username** (String) The username of the account.


<a id="nestedblock--relations"></a>
### Nested Schema for `relations`

Required:

- **value** (String) The name of the person the user is related to.

Optional:

- **custom_type** (String) If the value of type is custom, this property containsthe custom type string.
- **type** (String) The type of relation.Acceptable values: `admin_assistant`, `assistant`, `brother`, `child`, `custom`,`domestic_partner`, `dotted_line_manager`, `exec_assistant`, `father`, `friend`,`manager`, `mother`, `parent`, `partner`, `referred_by`, `relative`, `sister`,`spouse`.


<a id="nestedblock--ssh_public_keys"></a>
### Nested Schema for `ssh_public_keys`

Required:

- **key** (String) An SSH public key.

Optional:

- **expiration_time_usec** (String) An expiration time in microseconds since epoch.

Read-only:

- **fingerprint** (String) A SHA-256 fingerprint of the SSH public key.


<a id="nestedblock--websites"></a>
### Nested Schema for `websites`

Required:

- **value** (String) The URL of the website.

Optional:

- **custom_type** (String) The custom type. Only used if the type is custom.
- **primary** (Boolean) If this is user's primary website or not.
- **type** (String) The type or purpose of the website. For example, a website could be labeledas home or blog. Alternatively, an entry can have a custom type.Custom types must have a customType value.Acceptable values: `app_install_page`, `blog`, `custom`, `ftp`, `home`, `home_page`, `other`, `profile`, `reservations`, `resume`, `work`.


