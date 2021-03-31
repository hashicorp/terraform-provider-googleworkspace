package googleworkspace

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	directory "google.golang.org/api/admin/directory/v1"
)

func diffSuppressEmails(k, old, new string, d *schema.ResourceData) bool {
	stateEmails, configEmails := d.GetChange("emails")

	// The primary email (and potentially a test email with the format <primary_email>.test-google-a.com,
	// if the domain is the primary domain) are auto-added to the email list, even if it's not configured that way.
	// Only show a diff if the other emails differ.
	subsetEmails := []interface{}{}

	primaryEmail := d.Get("primary_email").(string)
	for _, se := range stateEmails.([]interface{}) {
		emailObj := se.(map[string]interface{})
		if emailObj["primary"].(bool) || emailObj["address"].(string) == fmt.Sprintf("%s.test-google-a.com", primaryEmail) {
			continue
		}

		subsetEmails = append(subsetEmails, se)
	}

	return reflect.DeepEqual(subsetEmails, configEmails.([]interface{}))
}

func resourceUser() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "User resource manages Google Workspace Users.",

		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceUserImport,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The unique ID for the user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"primary_email": {
				Description: "The user's primary email address. The primaryEmail must be unique and cannot be an alias " +
					"of another user.",
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Description: "Stores the password for the user account. A password can contain any combination of " +
					"ASCII characters. A minimum of 8 characters is required. The maximum length is 100 characters. " +
					"As the API does not return the value of password, this field is write-only, " +
					"and the value stored in the state will be what is provided in the configuration.",
				Type:             schema.TypeString,
				Required:         true,
				Sensitive:        true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(8, 100)),
			},
			"hash_function": {
				Description: "Stores the hash format of the password property. We recommend sending the password " +
					"property value as a base 16 bit hexadecimal-encoded hash value. Set the hashFunction values " +
					"as either the SHA-1, MD5, or crypt hash format.",
				Type:     schema.TypeString,
				Optional: true,
			},
			"is_admin": {
				Description: "Indicates a user with super admininistrator privileges.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"is_delegated_admin": {
				Description: "Indicates if the user is a delegated administrator.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"agreed_to_terms": {
				Description: "This property is true if the user has completed an initial login and accepted the " +
					"Terms of Service agreement.",
				Type:     schema.TypeBool,
				Computed: true,
			},
			"suspended": {
				Description: "Indicates if user is suspended.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"change_password_at_next_login": {
				Description: "Indicates if the user is forced to change their password at next login. This setting " +
					"doesn't apply when the user signs in via a third-party identity provider.",
				Type:     schema.TypeBool,
				Optional: true,
			},
			"ip_allowlist": {
				Description: "If true, the user's IP address is added to the allow list.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"name": {
				Description: "Holds the given and family names of the user, and the read-only fullName value. " +
					"The maximum number of characters in the givenName and in the familyName values is 60. " +
					"In addition, name values support unicode/UTF-8 characters, and can contain spaces, letters (a-z), " +
					"numbers (0-9), dashes (-), forward slashes (/), and periods (.). " +
					"Maximum allowed data size for this field is 1Kb.",
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"full_name": {
							Description: "The user's full name formed by concatenating the first and last name values.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"family_name": {
							Description:      "The user's last name.",
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 60)),
						},
						"given_name": {
							Description:      "The user's first name.",
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 60)),
						},
					},
				},
			},
			"etag": {
				Description: "ETag of the resource.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"emails": {
				Description:      "A list of the user's email addresses. The maximum allowed data size is 10Kb.",
				Type:             schema.TypeList,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: diffSuppressEmails,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Description: "The user's email address. Also serves as the email ID. " +
								"This value can be the user's primary email address or an alias.",
							Type:     schema.TypeString,
							Optional: true,
						},
						"custom_type": {
							Description: "If the value of type is custom, this property contains " +
								"the custom type string.",
							Type:     schema.TypeString,
							Optional: true,
						},
						"primary": {
							Description: "Indicates if this is the user's primary email. " +
								"Only one entry can be marked as primary.",
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"type": {
							Description: "The type of the email account. " +
								"Acceptable values: `custom`, `home`, `other`, `work`.",
							Type:     schema.TypeString,
							Optional: true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"custom", "home", "other", "work"}, false),
							),
						},
					},
				},
			},
			"external_ids": {
				Description: "A list of external IDs for the user, such as an employee or network ID. " +
					"The maximum allowed data size is 2Kb.",
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_type": {
							Description: "If the value of type is custom, this property contains " +
								"the custom type string.",
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Description: "The type of the email account. " +
								"Acceptable values: `custom`, `home`, `other`, `work`.",
							Type:     schema.TypeString,
							Optional: true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"custom", "home", "other", "work"}, false),
							),
						},
						"value": {
							Description: "The value of the ID.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"relations": {
				Description: "A list of the user's relationships to other users. " +
					"The maximum allowed data size for this field is 2Kb.",
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_type": {
							Description: "If the value of type is custom, this property contains " +
								"the custom type string.",
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Description: "The type of relation. " +
								"Acceptable values: `admin_assistant`, `assistant`, `brother`, `child`, `custom`, " +
								"`domestic_partner`, `dotted_line_manager`, `exec_assistant`, `father`, `friend`, " +
								"`manager`, `mother`, `parent`, `partner`, `referred_by`, `relative`, `sister`, " +
								"`spouse`.",
							Type:     schema.TypeString,
							Optional: true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"admin_assistant", "assistant", "brother", "child",
									"custom", "domestic_partner", "dotted_line_manager", "exec_assistant", "father",
									"friend", "manager", "mother", "parent", "partner", "referred_by", "relative",
									"sister"}, false),
							),
						},
						"value": {
							Description: "The name of the person the user is related to.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"aliases": {
				Description: "asps.list of the user's alias email addresses.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"is_mailbox_setup": {
				Description: "Indicates if the user's Google mailbox is created. This property is only applicable " +
					"if the user has been assigned a Gmail license.",
				Type:     schema.TypeBool,
				Computed: true,
			},
			"customer_id": {
				Description: "The customer ID to retrieve all account users. You can use the alias my_customer to " +
					"represent your account's customerId. As a reseller administrator, you can use the resold " +
					"customer account's customerId. To get a customerId, use the account's primary domain in the " +
					"domain parameter of a users.list request.",
				Type:     schema.TypeString,
				Computed: true,
			},
			// TODO: (mbang) AtLeastOneOf (https://github.com/hashicorp/terraform-plugin-sdk/issues/470)
			"addresses": {
				Description: "A list of the user's addresses. The maximum allowed data size is 10Kb.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"country": {
							Description: "Country",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"country_code": {
							Description: "The country code. Uses the ISO 3166-1 standard.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"custom_type": {
							Description: "If the address type is custom, this property contains the custom value.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"extended_address": {
							Description: "For extended addresses, such as an address that includes a sub-region.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"formatted": {
							Description: "A full and unstructured postal address. This is not synced with the " +
								"structured address fields.",
							Type:     schema.TypeString,
							Optional: true,
						},
						"locality": {
							Description: "The town or city of the address.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"po_box": {
							Description: "The post office box, if present.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"postal_code": {
							Description: "The ZIP or postal code, if applicable.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"primary": {
							Description: "If this is the user's primary address. The addresses list may contain " +
								"only one primary address.",
							Type:     schema.TypeBool,
							Optional: true,
						},
						"region": {
							Description: "The abbreviated province or state.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"source_is_structured": {
							Description: "Indicates if the user-supplied address was formatted. " +
								"Formatted addresses are not currently supported.",
							Type:     schema.TypeBool,
							Optional: true,
						},
						"street_address": {
							Description: "The street address, such as 1600 Amphitheatre Parkway. " +
								"Whitespace within the string is ignored; however, newlines are significant.",
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Description: "The address type. " +
								"Acceptable values: `custom`, `home`, `other`, `work`.",
							Type:     schema.TypeString,
							Optional: true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"custom", "home", "other", "work"}, false),
							),
						},
					},
				},
			},
			"organizations": {
				Description: "A list of organizations the user belongs to. The maximum allowed data size is 10Kb.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cost_center": {
							Description: "The cost center of the user's organization.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"custom_type": {
							Description: "If the value of type is custom, this property contains the custom value.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"department": {
							Description: "Specifies the department within the organization, such as sales or engineering.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"description": {
							Description: "The description of the organization.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"domain": {
							Description: "The domain the organization belongs to.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"full_time_equivalent": {
							Description: "The full-time equivalent millipercent within the organization (100000 = 100%)",
							Type:        schema.TypeInt,
							Optional:    true,
						},
						"location": {
							Description: "The physical location of the organization. " +
								"This does not need to be a fully qualified address.",
							Type:     schema.TypeString,
							Optional: true,
						},
						"name": {
							Description: "The name of the organization.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"primary": {
							Description: "Indicates if this is the user's primary organization. A user may only have " +
								"one primary organization.",
							Type:     schema.TypeBool,
							Optional: true,
						},
						"symbol": {
							Description: "Text string symbol of the organization. For example, " +
								"the text symbol for Google is GOOG.",
							Type:     schema.TypeString,
							Optional: true,
						},
						"title": {
							Description: "The user's title within the organization. For example, member or engineer.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"type": {
							Description: "The type of organization. " +
								"Acceptable values: `domain_only`, `school`, `unknown`, `work`.",
							Type:     schema.TypeString,
							Optional: true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"domain_only", "school", "unknown", "work"}, false),
							),
						},
					},
				},
			},
			"last_login_time": {
				Description: "The last time the user logged into the user's account. The value is in ISO 8601 date " +
					"and time format. The time is the complete date plus hours, minutes, and seconds " +
					"in the form YYYY-MM-DDThh:mm:ssTZD. For example, 2010-04-05T17:30:04+01:00.",
				Type:     schema.TypeString,
				Computed: true,
			},
			"phones": {
				Description: "Holds the given and family names of the user, and the read-only fullName value. " +
					"The maximum number of characters in the givenName and in the familyName values is 60. " +
					"In addition, name values support unicode/UTF-8 characters, and can contain spaces, letters (a-z), " +
					"numbers (0-9), dashes (-), forward slashes (/), and periods (.). " +
					"Maximum allowed data size for this field is 1Kb.",
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_type": {
							Description: "If the value of type is custom, this property contains the custom type.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"primary": {
							Description: "Indicates if this is the user's primary phone number. " +
								"A user may only have one primary phone number.",
							Type:     schema.TypeBool,
							Optional: true,
						},
						"type": {
							Description: "The type of phone number. " +
								"Acceptable values: `assistant`, `callback`, `car`, `company_main` " +
								", `custom`, `grand_central`, `home`, `home_fax`, `isdn`, `main`, `mobile`, `other`, " +
								"`other_fax`, `pager`, `radio`, `telex`, `tty_tdd`, `work`, `work_fax`, `work_mobile`, " +
								"`work_pager`.",
							Type:     schema.TypeString,
							Optional: true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"assistant", "callback", "car", "company_main",
									"custom", "grand_central", "home", "home_fax", "isdn", "main", "mobile", "other",
									"other_fax", "pager", "radio", "telex", "tty_tdd", "work", "work_fax",
									"work_mobile", "work_pager"}, false),
							),
						},
						"value": {
							Description: "A human-readable phone number. It may be in any telephone number format.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"suspension_reason": {
				Description: "Has the reason a user account is suspended either by the administrator or by Google at " +
					"the time of suspension. The property is returned only if the suspended property is true.",
				Type:     schema.TypeString,
				Computed: true,
			},
			"thumbnail_photo_url": {
				Description: "Photo Url of the user.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"languages": {
				Description: "A list of the user's languages. The maximum allowed data size is 1Kb.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_language": {
							Description: "Other language. A user can provide their own language name if there is no " +
								"corresponding Google III language code. If this is set, LanguageCode can't be set.",
							Type:     schema.TypeString,
							Optional: true,
							// TODO: (mbang) https://github.com/hashicorp/terraform-plugin-sdk/issues/470
							//ExactlyOneOf: []string{"custom_language", "language_code"},
						},
						"language_code": {
							Description: "Language Code. Should be used for storing Google III LanguageCode string " +
								"representation for language. Illegal values cause SchemaException.",
							Type:     schema.TypeString,
							Optional: true,
							// TODO: (mbang) https://github.com/hashicorp/terraform-plugin-sdk/issues/470
							//ExactlyOneOf: []string{"custom_language", "language_code"},
						},
					},
				},
			},
			// TODO: (mbang) AtLeastOneOf (https://github.com/hashicorp/terraform-plugin-sdk/issues/470)
			"posix_accounts": {
				Description: "A list of POSIX account information for the user.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_id": {
							Description: "A POSIX account field identifier.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"gecos": {
							Description: "The GECOS (user information) for this account.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"gid": {
							Description: "The default group ID.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"home_directory": {
							Description: "The path to the home directory for this account.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"operating_system_type": {
							Description: "The operating system type for this account. " +
								"Acceptable values: `linux`, `unspecified`, `windows`.",
							Type:     schema.TypeString,
							Optional: true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"linux", "unspecified", "windows"}, false),
							),
						},
						"primary": {
							Description: "If this is user's primary account within the SystemId.",
							Type:        schema.TypeBool,
							Optional:    true,
						},
						"shell": {
							Description: "The path to the login shell for this account.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"system_id": {
							Description: "System identifier for which account Username or Uid apply to.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"uid": {
							Description: "The POSIX compliant user ID.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"username": {
							Description: "The username of the account.",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
			"creation_time": {
				Description: "The time the user's account was created. The value is in ISO 8601 date and time format. " +
					"The time is the complete date plus hours, minutes, and seconds in the form " +
					"YYYY-MM-DDThh:mm:ssTZD. For example, 2010-04-05T17:30:04+01:00.",
				Type:     schema.TypeString,
				Computed: true,
			},
			"non_editable_aliases": {
				Description: "asps.list of the user's non-editable alias email addresses. These are typically outside " +
					"the account's primary domain or sub-domain.",
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ssh_public_keys": {
				Description: "A list of SSH public keys. The maximum allowed data size is 10Kb.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"expiration_time_usec": {
							Description: "An expiration time in microseconds since epoch.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"fingerprint": {
							Description: "A SHA-256 fingerprint of the SSH public key.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"key": {
							Description: "An SSH public key.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"websites": {
				Description: "A list of the user's websites. The maximum allowed data size is 2Kb.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_type": {
							Description: "The custom type. Only used if the type is custom.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"primary": {
							Description: "If this is user's primary website or not.",
							Type:        schema.TypeBool,
							Optional:    true,
						},
						"type": {
							Description: "The type or purpose of the website. For example, a website could be labeled " +
								"as home or blog. Alternatively, an entry can have a custom type " +
								"Custom types must have a customType value. " +
								"Acceptable values: `app_install_page`, `blog`, `custom`, `ftp` " +
								", `home`, `home_page`, `other`, `profile`, `reservations`, `resume`, `work`.",
							Type:     schema.TypeString,
							Optional: true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"app_install_page", "blog", "custom", "ftp",
									"home", "home_page", "other", "profile", "reservations", "resume", "work"},
									false),
							),
						},
						"value": {
							Description: "The URL of the website.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"locations": {
				Description: "A list of the user's locations. The maximum allowed data size is 10Kb.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"area": {
							Description: "Textual location. This is most useful for display purposes to concisely " +
								"describe the location. For example, Mountain View, CA or Near Seattle.",
							Type:     schema.TypeString,
							Optional: true,
						},
						"building_id": {
							Description: "Building identifier.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"custom_type": {
							Description: "If the location type is custom, this property contains the custom value.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"desk_code": {
							Description: "Most specific textual code of individual desk location.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"floor_name": {
							Description: "Floor name/number.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"floor_section": {
							Description: "Floor section. More specific location within the floor. For example, " +
								"if a floor is divided into sections A, B, and C, this field would identify one " +
								"of those values.",
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Description: "The location type. " +
								"Acceptable values: `custom`, `default`, `desk`",
							Type:     schema.TypeString,
							Optional: true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"custom", "default", "desk"},
									false),
							),
						},
					},
				},
			},
			// IncludeInGlobalAddressList is not being sent in the request with the admin SDK, so leaving this out for now
			"include_in_global_address_list": {
				Description: "Indicates if the user's profile is visible in the Google Workspace global address list " +
					"when the contact sharing feature is enabled for the domain.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"keywords": {
				Description: "A list of the user's keywords. The maximum allowed data size is 1Kb.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_type": {
							Description: "Custom Type.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"type": {
							Description: "Each entry can have a type which indicates standard type of that entry. " +
								"For example, keyword could be of type occupation or outlook. In addition to the " +
								"standard type, an entry can have a custom type and can give it any name. Such types " +
								"should have the CUSTOM value as type and also have a customType value. " +
								"Acceptable values: `custom`, `mission`, `occupation`, `outlook`",
							Type:     schema.TypeString,
							Optional: true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"custom", "mission", "occupation", "outlook"},
									false),
							),
						},
						"value": {
							Description: "Keyword.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"deletion_time": {
				Description: "The time the user's account was deleted. The value is in ISO 8601 date and time format " +
					"The time is the complete date plus hours, minutes, and seconds in the form YYYY-MM-DDThh:mm:ssTZD. " +
					"For example 2010-04-05T17:30:04+01:00.",
				Type:     schema.TypeString,
				Computed: true,
			},
			// "gender" is not included in the GET response, currently, so leaving this out for now
			"thumbnail_photo_etag": {
				Description: "ETag of the user's photo",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ims": {
				Description: "The user's Instant Messenger (IM) accounts. A user account can have multiple ims " +
					"properties. But, only one of these ims properties can be the primary IM contact. " +
					"The maximum allowed data size is 2Kb.",
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_protocol": {
							Description: "If the protocol value is custom_protocol, this property holds the custom " +
								"protocol's string.",
							Type:     schema.TypeString,
							Optional: true,
						},
						"custom_type": {
							Description: "If the IM type is custom, this property holds the custom type string.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"im": {
							Description: "The user's IM network ID.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"primary": {
							Description: "If this is the user's primary IM. " +
								"Only one entry in the IM list can have a value of true.",
							Type:     schema.TypeBool,
							Optional: true,
						},
						"protocol": {
							Description: "An IM protocol identifies the IM network. " +
								"The value can be a custom network or the standard network. " +
								"Acceptable values: `aim`, `custom_protocol`, `gtalk`, `icq`, `jabber`, " +
								"`msn`, `net_meeting`, `qq`, `skype`, `yahoo`.",
							Type:     schema.TypeString,
							Optional: true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"aim", "custom_protocol", "gtalk", "icq",
									"jabber", "msn", "net_meeting", "qq", "skype", "yahoo"}, false),
							),
						},
						"type": {
							Description: "Acceptable values: `home`, `callback`, `other`, `work`.",
							Type:        schema.TypeString,
							Optional:    true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"custom", "home", "other", "work"}, false),
							),
						},
					},
				},
			},
			// TODO: (mbang) custom schemas
			"is_enrolled_in_2_step_verification": {
				Description: "Is enrolled in 2-step verification.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"is_enforced_in_2_step_verification": {
				Description: "Is 2-step verification enforced.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"archived": {
				Description: "Indicates if user is archived.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"org_unit_path": {
				Description: "The full path of the parent organization associated with the user. " +
					"If the parent organization is the top-level, it is represented as a forward slash (/).",
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"recovery_email": {
				Description: "Recovery email of the user.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"recovery_phone": {
				Description: "Recovery phone of the user. The phone number must be in the E.164 format, " +
					"starting with the plus sign (+). Example: +16506661212.",
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	primaryEmail := d.Get("primary_email").(string)
	log.Printf("[DEBUG] Creating User %q: %#v", d.Id(), primaryEmail)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	usersService, diags := GetUsersService(directoryService)
	if diags.HasError() {
		return diags
	}

	userObj := directory.User{
		PrimaryEmail:               primaryEmail,
		Password:                   d.Get("password").(string),
		HashFunction:               d.Get("hash_function").(string),
		Suspended:                  d.Get("suspended").(bool),
		ChangePasswordAtNextLogin:  d.Get("change_password_at_next_login").(bool),
		IpWhitelisted:              d.Get("ip_allowlist").(bool),
		Name:                       expandName(d.Get("name")),
		Emails:                     expandInterfaceObjects(d.Get("emails")),
		ExternalIds:                expandInterfaceObjects(d.Get("external_ids")),
		Relations:                  expandInterfaceObjects(d.Get("relations")),
		Addresses:                  expandInterfaceObjects(d.Get("addresses")),
		Organizations:              expandInterfaceObjects(d.Get("organizations")),
		Phones:                     expandInterfaceObjects(d.Get("phones")),
		Languages:                  expandInterfaceObjects(d.Get("languages")),
		PosixAccounts:              expandInterfaceObjects(d.Get("posix_accounts")),
		SshPublicKeys:              expandInterfaceObjects(d.Get("ssh_public_keys")),
		Websites:                   expandInterfaceObjects(d.Get("websites")),
		Locations:                  expandInterfaceObjects(d.Get("locations")),
		IncludeInGlobalAddressList: d.Get("include_in_global_address_list").(bool),
		Keywords:                   expandInterfaceObjects(d.Get("keywords")),
		Ims:                        expandInterfaceObjects(d.Get("ims")),
		Archived:                   d.Get("archived").(bool),
		OrgUnitPath:                d.Get("org_unit_path").(string),
		RecoveryEmail:              d.Get("recovery_email").(string),
		RecoveryPhone:              d.Get("recovery_phone").(string),

		ForceSendFields: []string{"IncludeInGlobalAddressList"},
	}

	user, err := usersService.Insert(&userObj).Do()
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(user.Id)

	if d.Get("is_admin").(bool) {
		makeAdminObj := directory.UserMakeAdmin{
			Status: d.Get("is_admin").(bool),
		}

		err = usersService.MakeAdmin(user.Id, &makeAdminObj).Do()
		if err != nil {
			return diag.FromErr(err)
		}
	}

	time.Sleep(15 * time.Second)

	log.Printf("[DEBUG] Finished creating User %q: %#v", d.Id(), primaryEmail)
	return resourceUserRead(ctx, d, meta)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	usersService, diags := GetUsersService(directoryService)
	if diags.HasError() {
		return diags
	}

	user, err := usersService.Get(d.Id()).Projection("full").Do()
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("id", user.Id)
	d.Set("primary_email", user.PrimaryEmail)
	// password and hash_function are not returned in the response, so set them to what we defined in the config
	d.Set("password", d.Get("password"))
	d.Set("hash_function", d.Get("hash_function"))
	d.Set("is_admin", user.IsAdmin)
	d.Set("is_delegated_admin", user.IsDelegatedAdmin)
	d.Set("agreed_to_terms", user.AgreedToTerms)
	d.Set("suspended", user.Suspended)
	d.Set("change_password_at_next_login", user.ChangePasswordAtNextLogin)
	d.Set("ip_allowlist", user.IpWhitelisted)
	d.Set("name", flattenName(user.Name))
	d.Set("emails", flattenAndSetInterfaceObjects(user.Emails))
	d.Set("external_ids", flattenAndSetInterfaceObjects(user.ExternalIds))
	d.Set("relations", flattenAndSetInterfaceObjects(user.Relations))
	d.Set("etag", user.Etag)
	d.Set("aliases", user.Aliases)
	d.Set("is_mailbox_setup", user.IsMailboxSetup)
	d.Set("customer_id", user.CustomerId)
	d.Set("addresses", flattenAndSetInterfaceObjects(user.Addresses))
	d.Set("organizations", flattenAndSetInterfaceObjects(user.Organizations))
	d.Set("last_login_time", user.LastLoginTime)
	d.Set("phones", flattenAndSetInterfaceObjects(user.Phones))
	d.Set("suspension_reason", user.SuspensionReason)
	d.Set("thumbnail_photo_url", user.ThumbnailPhotoUrl)
	d.Set("languages", flattenAndSetInterfaceObjects(user.Languages))
	d.Set("posix_accounts", flattenAndSetInterfaceObjects(user.PosixAccounts))
	d.Set("creation_time", user.CreationTime)
	d.Set("non_editable_aliases", user.NonEditableAliases)
	d.Set("ssh_public_keys", flattenAndSetInterfaceObjects(user.SshPublicKeys))
	d.Set("websites", flattenAndSetInterfaceObjects(user.Websites))
	d.Set("locations", flattenAndSetInterfaceObjects(user.Locations))
	d.Set("include_in_global_address_list", user.IncludeInGlobalAddressList)
	d.Set("keywords", flattenAndSetInterfaceObjects(user.Keywords))
	d.Set("deletion_time", user.DeletionTime)
	d.Set("thumbnail_photo_etag", user.ThumbnailPhotoEtag)
	d.Set("ims", flattenAndSetInterfaceObjects(user.Ims))
	d.Set("is_enrolled_in_2_step_verification", user.IsEnrolledIn2Sv)
	d.Set("is_enforced_in_2_step_verification", user.IsEnforcedIn2Sv)
	d.Set("archived", user.Archived)
	d.Set("org_unit_path", user.OrgUnitPath)
	d.Set("recovery_email", user.RecoveryEmail)
	d.Set("recovery_phone", user.RecoveryPhone)

	d.SetId(user.Id)

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	primaryEmail := d.Get("primary_email").(string)
	log.Printf("[DEBUG] Updating User %q: %#v", d.Id(), primaryEmail)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	usersService, diags := GetUsersService(directoryService)
	if diags.HasError() {
		return diags
	}

	userObj := directory.User{}

	// Strings

	if d.HasChange("primary_email") {
		userObj.PrimaryEmail = primaryEmail
	}

	if d.HasChange("password") {
		userObj.Password = d.Get("password").(string)
	}

	if d.HasChange("hash_function") {
		userObj.HashFunction = d.Get("hash_function").(string)
	}

	if d.HasChange("org_unit_path") {
		userObj.OrgUnitPath = d.Get("org_unit_path").(string)
	}

	if d.HasChange("recovery_email") {
		userObj.RecoveryEmail = d.Get("recovery_email").(string)
	}

	if d.HasChange("recovery_phone") {
		userObj.RecoveryPhone = d.Get("recovery_phone").(string)
	}

	// Booleans (need to send ForceNewFields)
	forceSendFields := []string{}

	if d.HasChange("suspended") {
		userObj.Suspended = d.Get("suspended").(bool)
		forceSendFields = append(forceSendFields, "Suspended")
	}

	if d.HasChange("change_password_at_next_login") {
		userObj.ChangePasswordAtNextLogin = d.Get("change_password_at_next_login").(bool)
		forceSendFields = append(forceSendFields, "ChangePasswordAtNextLogin")
	}

	if d.HasChange("ip_allowlist") {
		userObj.IpWhitelisted = d.Get("ip_allowlist").(bool)
		forceSendFields = append(forceSendFields, "IpWhitelisted")
	}

	if d.HasChange("include_in_global_address_list") {
		userObj.IncludeInGlobalAddressList = d.Get("include_in_global_address_list").(bool)
		forceSendFields = append(forceSendFields, "IncludeInGlobalAddressList")
	}

	if d.HasChange("archived") {
		userObj.Archived = d.Get("archived").(bool)
		forceSendFields = append(forceSendFields, "Archived")
	}

	userObj.ForceSendFields = forceSendFields

	// Nested Objects

	if d.HasChange("name") {
		userObj.Name = expandName(d.Get("name"))
	}

	if d.HasChange("emails") {
		emails := expandInterfaceObjects(d.Get("emails"))
		userObj.Emails = emails
	}

	if d.HasChange("external_ids") {
		externalIds := expandInterfaceObjects(d.Get("external_ids"))
		userObj.ExternalIds = externalIds
	}

	if d.HasChange("relations") {
		emails := expandInterfaceObjects(d.Get("relations"))
		userObj.Relations = emails
	}

	if d.HasChange("addresses") {
		addresses := expandInterfaceObjects(d.Get("addresses"))
		userObj.Addresses = addresses
	}

	if d.HasChange("organizations") {
		organizations := expandInterfaceObjects(d.Get("organizations"))
		userObj.Organizations = organizations
	}

	if d.HasChange("phones") {
		phones := expandInterfaceObjects(d.Get("phones"))
		userObj.Phones = phones
	}

	if d.HasChange("languages") {
		languages := expandInterfaceObjects(d.Get("languages"))
		userObj.Languages = languages
	}

	if d.HasChange("posix_accounts") {
		posixAccounts := expandInterfaceObjects(d.Get("posix_accounts"))
		userObj.PosixAccounts = posixAccounts
	}

	if d.HasChange("ssh_public_keys") {
		sshPublicKeys := expandInterfaceObjects(d.Get("ssh_public_keys"))
		userObj.SshPublicKeys = sshPublicKeys
	}

	if d.HasChange("websites") {
		websites := expandInterfaceObjects(d.Get("websites"))
		userObj.Websites = websites
	}

	if d.HasChange("locations") {
		locations := expandInterfaceObjects(d.Get("locations"))
		userObj.Locations = locations
	}

	if d.HasChange("keywords") {
		keywords := expandInterfaceObjects(d.Get("keywords"))
		userObj.Keywords = keywords
	}

	if d.HasChange("ims") {
		ims := expandInterfaceObjects(d.Get("ims"))
		userObj.Ims = ims
	}

	if &userObj != new(directory.User) {
		user, err := usersService.Update(d.Id(), &userObj).Do()
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(user.Id)
	}

	if d.HasChange("is_admin") {
		makeAdminObj := directory.UserMakeAdmin{
			Status:          d.Get("is_admin").(bool),
			ForceSendFields: []string{"Status"},
		}

		err := usersService.MakeAdmin(d.Id(), &makeAdminObj).Do()
		if err != nil {
			return diag.FromErr(err)
		}
	}

	time.Sleep(60 * time.Second)

	log.Printf("[DEBUG] Finished updating User %q: %#v", d.Id(), primaryEmail)

	return resourceUserRead(ctx, d, meta)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	primaryEmail := d.Get("primary_email").(string)
	log.Printf("[DEBUG] Deleting User %q: %#v", d.Id(), primaryEmail)

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	usersService, diags := GetUsersService(directoryService)
	if diags.HasError() {
		return diags
	}

	err := usersService.Delete(d.Id()).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished deleting User %q: %#v", d.Id(), primaryEmail)

	return diags
}

func resourceUserImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}

// Expand functions

func expandName(v interface{}) *directory.UserName {
	nameObj := v.([]interface{})

	if len(nameObj) == 0 {
		return nil
	}

	userNameObj := directory.UserName{
		FamilyName: nameObj[0].(map[string]interface{})["family_name"].(string),
		GivenName:  nameObj[0].(map[string]interface{})["given_name"].(string),
	}
	return &userNameObj
}

// User type has many nested interfaces, we can pass them to the API as is
// only each field name needs to be camel case rather than snake case.
func expandInterfaceObjects(v interface{}) []interface{} {
	objList := v.([]interface{})
	if objList == nil || len(objList) == 0 {
		return nil
	}

	newObjList := []interface{}{}

	for _, o := range objList {
		obj := o.(map[string]interface{})
		for k, v := range obj {
			if strings.Contains(k, "_") {
				delete(obj, k)

				// In the case that the field is not set, don't send it to the API
				if v == "" {
					continue
				}

				obj[SnakeToCamel(k)] = v
			}
		}
		newObjList = append(newObjList, obj)
	}

	return newObjList
}

// Flatten functions

func flattenName(userNameObj *directory.UserName) interface{} {
	nameObj := []map[string]interface{}{}

	if userNameObj != nil {
		nameObj = append(nameObj, map[string]interface{}{
			"family_name": userNameObj.FamilyName,
			"full_name":   userNameObj.FullName,
			"given_name":  userNameObj.GivenName,
		})
	}

	return nameObj
}

// User type has many nested interfaces, we can set was was returned from the API as is
// only the field names need to be snake case rather than the camel case that is returned
func flattenAndSetInterfaceObjects(objList interface{}) interface{} {
	if objList == nil || len(objList.([]interface{})) == 0 {
		return nil
	}

	newObjList := []map[string]interface{}{}

	for _, o := range objList.([]interface{}) {
		obj := o.(map[string]interface{})
		for k, v := range obj {
			delete(obj, k)
			obj[CameltoSnake(k)] = v
		}

		newObjList = append(newObjList, obj)
	}

	return newObjList
}
