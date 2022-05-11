package googleworkspace

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-googleworkspace-pf/internal/model"
	directory "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
	"log"
	"net/mail"
	"reflect"
	"strconv"
	"time"
)

type resourceUserType struct{}

// GetSchema User Resource
func (r resourceUserType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: "User resource manages Google Workspace Users. User resides " +
			"under the `https://www.googleapis.com/auth/admin.directory.user` client scope.",
		Attributes: map[string]tfsdk.Attribute{
			"primary_email": {
				Description: "The user's primary email address. The primaryEmail must be unique and cannot be an alias " +
					"of another user.",
				Type:     types.StringType,
				Required: true,
			},
			"password": {
				Description: "Stores the password for the user account. A password can contain any combination of " +
					"ASCII characters. A minimum of 8 characters is required. The maximum length is 100 characters. " +
					"As the API does not return the value of password, this field is write-only, and the value stored " +
					"in the state will be what is provided in the configuration. The field is required on create and will " +
					"be empty on import.",
				Type:      types.StringType,
				Required:  true,
				Sensitive: true,
				Validators: []tfsdk.AttributeValidator{
					StringLenBetweenValidator{
						Min: 8,
						Max: 100,
					},
				},
			},
			"hash_function": {
				Description: "Stores the hash format of the password property. We recommend sending the password " +
					"property value as a base 16 bit hexadecimal-encoded hash value. Set the hashFunction values " +
					"as either the SHA-1, MD5, or crypt hash format.",
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
			},
			"is_admin": {
				Description: "Indicates a user with super administrator privileges.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
			},
			"is_delegated_admin": {
				Description: "Indicates if the user is a delegated administrator.",
				Type:        types.BoolType,
				Computed:    true,
			},
			"agreed_to_terms": {
				Description: "This property is true if the user has completed an initial login and accepted the " +
					"Terms of Service agreement.",
				Type:     types.BoolType,
				Computed: true,
			},
			"suspended": {
				Description: "Indicates if user is suspended.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.Bool{Value: false},
					},
				},
			},
			"change_password_at_next_login": {
				Description: "Indicates if the user is forced to change their password at next login. This setting " +
					"doesn't apply when the user signs in via a third-party identity provider.",
				Type:     types.BoolType,
				Optional: true,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.Bool{Value: false},
					},
				},
			},
			"ip_allowlist": {
				Description: "If true, the user's IP address is added to the allow list.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.Bool{Value: false},
					},
				},
			},
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
			"name": {
				Description: "Holds the given and family names of the user, and the read-only fullName value. " +
					"The maximum number of characters in the givenName and in the familyName values is 60. " +
					"In addition, name values support unicode/UTF-8 characters, and can contain spaces, letters (a-z), " +
					"numbers (0-9), dashes (-), forward slashes (/), and periods (.). " +
					"Maximum allowed data size for this field is 1Kb.",
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"full_name": {
						Description: "The user's full name formed by concatenating the first and last name values.",
						Type:        types.StringType,
						Computed:    true,
					},
					"family_name": {
						Description: "The user's last name.",
						Type:        types.StringType,
						Required:    true,
						Validators: []tfsdk.AttributeValidator{
							StringLenBetweenValidator{
								Min: 1,
								Max: 60,
							},
						},
					},
					"given_name": {
						Description: "The user's first name.",
						Type:        types.StringType,
						Required:    true,
						Validators: []tfsdk.AttributeValidator{
							StringLenBetweenValidator{
								Min: 1,
								Max: 60,
							},
						},
					},
				}),
				Optional: true,
			},
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
			"emails": {
				Description: "A list of the user's email addresses. The maximum allowed data size is 10Kb.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					UserEmailsModifier{},
				},
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"address": {
						Description: "The user's email address. Also serves as the email ID. " +
							"This value can be the user's primary email address or an alias.",
						Type:     types.StringType,
						Required: true,
					},
					"custom_type": {
						Description: "If the value of type is custom, this property contains " +
							"the custom type string.",
						Type:     types.StringType,
						Optional: true,
					},
					"primary": {
						Description: "Indicates if this is the user's primary email. " +
							"Only one entry can be marked as primary.",
						Type:     types.BoolType,
						Optional: true,
						Computed: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							DefaultModifier{
								DefaultValue: types.Bool{Value: false},
							},
						},
					},
					"type": {
						Description: "The type of the email account. " +
							"Acceptable values: `custom`, `home`, `other`, `work`.",
						Type:     types.StringType,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							StringInSliceValidator{
								Options: []string{"custom", "home", "other", "work"},
							},
						},
					},
				}, tfsdk.ListNestedAttributesOptions{}),
			},
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
			"external_ids": {
				Description: "A list of external IDs for the user, such as an employee or network ID. " +
					"The maximum allowed data size is 2Kb.",
				Optional: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"custom_type": {
						Description: "If the external ID type is custom, this property contains the custom value and " +
							"must be set.",
						Type:     types.StringType,
						Optional: true,
					},
					"type": {
						Description: "The type of external ID. If set to custom, customType must also be set. " +
							"Acceptable values: `account`, `custom`, `customer`, `login_id`, `network`, `organization`.",
						Type:     types.StringType,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							StringInSliceValidator{
								Options: []string{"account", "custom", "customer", "login_id",
									"network", "organization"},
							},
						},
					},
					"value": {
						Description: "The value of the ID.",
						Type:        types.StringType,
						Required:    true,
					},
				}, tfsdk.ListNestedAttributesOptions{}),
			},
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
			"relations": {
				Description: "A list of the user's relationships to other users. " +
					"The maximum allowed data size for this field is 2Kb.",
				Optional: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"custom_type": {
						Description: "If the value of type is custom, this property contains " +
							"the custom type string.",
						Type:     types.StringType,
						Optional: true,
					},
					"type": {
						Description: "The type of relation." +
							"Acceptable values: `admin_assistant`, `assistant`, `brother`, `child`, `custom`, " +
							"`domestic_partner`, `dotted_line_manager`, `exec_assistant`, `father`, `friend`, " +
							"`manager`, `mother`, `parent`, `partner`, `referred_by`, `relative`, `sister`, " +
							"`spouse`.",
						Type:     types.StringType,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							StringInSliceValidator{
								Options: []string{"admin_assistant", "assistant", "brother", "child",
									"custom", "domestic_partner", "dotted_line_manager", "exec_assistant", "father",
									"friend", "manager", "mother", "parent", "partner", "referred_by", "relative",
									"sister"},
							},
						},
					},
					"value": {
						Description: "The name of the person the user is related to.",
						Type:        types.StringType,
						Required:    true,
					},
				}, tfsdk.ListNestedAttributesOptions{}),
			},
			"aliases": {
				Description: "asps.list of the user's alias email addresses.",
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
			"is_mailbox_setup": {
				Description: "Indicates if the user's Google mailbox is created. This property is only applicable " +
					"if the user has been assigned a Gmail license.",
				Type:     types.BoolType,
				Computed: true,
			},
			"customer_id": {
				Description: "The customer ID to retrieve all account users. You can use the alias my_customer to " +
					"represent your account's customerId. As a reseller administrator, you can use the resold " +
					"customer account's customerId. To get a customerId, use the account's primary domain in the " +
					"domain parameter of a users.list request.",
				Type:     types.StringType,
				Computed: true,
			},
			// And add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
			"addresses": {
				Description: "A list of the user's addresses. The maximum allowed data size is 10Kb.",
				Optional:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"country": {
						Description: "Country",
						Type:        types.StringType,
						Optional:    true,
					},
					"country_code": {
						Description: "The country code. Uses the ISO 3166-1 standard.",
						Type:        types.StringType,
						Optional:    true,
					},
					"custom_type": {
						Description: "If the address type is custom, this property contains the custom value.",
						Type:        types.StringType,
						Optional:    true,
					},
					"extended_address": {
						Description: "For extended addresses, such as an address that includes a sub-region.",
						Type:        types.StringType,
						Optional:    true,
					},
					"formatted": {
						Description: "A full and unstructured postal address. This is not synced with the " +
							"structured address fields.",
						Type:     types.StringType,
						Optional: true,
					},
					"locality": {
						Description: "The town or city of the address.",
						Type:        types.StringType,
						Optional:    true,
					},
					"po_box": {
						Description: "The post office box, if present.",
						Type:        types.StringType,
						Optional:    true,
					},
					"postal_code": {
						Description: "The ZIP or postal code, if applicable.",
						Type:        types.StringType,
						Optional:    true,
					},
					"primary": {
						Description: "If this is the user's primary address. The addresses list may contain " +
							"only one primary address.",
						Type:     types.BoolType,
						Optional: true,
					},
					"region": {
						Description: "The abbreviated province or state.",
						Type:        types.StringType,
						Optional:    true,
					},
					"source_is_structured": {
						Description: "Indicates if the user-supplied address was formatted. " +
							"Formatted addresses are not currently supported.",
						Type:     types.BoolType,
						Optional: true,
					},
					"street_address": {
						Description: "The street address, such as 1600 Amphitheatre Parkway. " +
							"Whitespace within the string is ignored; however, newlines are significant.",
						Type:     types.StringType,
						Optional: true,
					},
					"type": {
						Description: "The address type." +
							"Acceptable values: `custom`, `home`, `other`, `work`.",
						Type:     types.StringType,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							StringInSliceValidator{
								Options: []string{"custom", "home", "other", "work"},
							},
						},
					},
				}, tfsdk.ListNestedAttributesOptions{}),
			},
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
			"organizations": {
				Description: "A list of organizations the user belongs to. The maximum allowed data size is 10Kb.",
				Optional:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"cost_center": {
						Description: "The cost center of the user's organization.",
						Type:        types.StringType,
						Optional:    true,
					},
					"custom_type": {
						Description: "If the value of type is custom, this property contains the custom value.",
						Type:        types.StringType,
						Optional:    true,
					},
					"department": {
						Description: "Specifies the department within the organization, such as sales or engineering.",
						Type:        types.StringType,
						Optional:    true,
					},
					"description": {
						Description: "The description of the organization.",
						Type:        types.StringType,
						Optional:    true,
					},
					"domain": {
						Description: "The domain the organization belongs to.",
						Type:        types.StringType,
						Optional:    true,
					},
					"full_time_equivalent": {
						Description: "The full-time equivalent millipercent within the organization (100000 = 100%)",
						Type:        types.StringType,
						Optional:    true,
					},
					"location": {
						Description: "The physical location of the organization. " +
							"This does not need to be a fully qualified address.",
						Type:     types.StringType,
						Optional: true,
					},
					"name": {
						Description: "The name of the organization.",
						Type:        types.StringType,
						Optional:    true,
					},
					"primary": {
						Description: "Indicates if this is the user's primary organization. A user may only have " +
							"one primary organization.",
						Type:     types.BoolType,
						Optional: true,
					},
					"symbol": {
						Description: "Text string symbol of the organization. For example, " +
							"the text symbol for Google is GOOG.",
						Type:     types.StringType,
						Optional: true,
					},
					"title": {
						Description: "The user's title within the organization. For example, member or engineer.",
						Type:        types.StringType,
						Optional:    true,
					},
					"type": {
						Description: "The type of organization." +
							"Acceptable values: `domain_only`, `school`, `unknown`, `work`.",
						Type:     types.StringType,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							StringInSliceValidator{
								Options: []string{"domain_only", "school", "unknown", "work"},
							},
						},
					},
				}, tfsdk.ListNestedAttributesOptions{}),
			},
			"last_login_time": {
				Description: "The last time the user logged into the user's account. The value is in ISO 8601 date " +
					"and time format. The time is the complete date plus hours, minutes, and seconds " +
					"in the form YYYY-MM-DDThh:mm:ssTZD. For example, 2010-04-05T17:30:04+01:00.",
				Type:     types.StringType,
				Computed: true,
			},
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
			"phones": {
				Description: "A list of the user's phone numbers. The maximum allowed data size is 1Kb.",
				Optional:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"custom_type": {
						Description: "If the phone number type is custom, this property contains the custom value " +
							"and must be set.",
						Type:     types.StringType,
						Optional: true,
					},
					"primary": {
						Description: "Indicates if this is the user's primary phone number. " +
							"A user may only have one primary phone number.",
						Type:     types.BoolType,
						Optional: true,
					},
					"type": {
						Description: "The type of phone number." +
							"Acceptable values: `assistant`, `callback`, `car`, `company_main` " +
							", `custom`, `grand_central`, `home`, `home_fax`, `isdn`, `main`, `mobile`, `other`, " +
							"`other_fax`, `pager`, `radio`, `telex`, `tty_tdd`, `work`, `work_fax`, `work_mobile`, " +
							"`work_pager`.",
						Type:     types.StringType,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							StringInSliceValidator{
								Options: []string{"assistant", "callback", "car", "company_main",
									"custom", "grand_central", "home", "home_fax", "isdn", "main", "mobile", "other",
									"other_fax", "pager", "radio", "telex", "tty_tdd", "work", "work_fax",
									"work_mobile", "work_pager"},
							},
						},
					},
					"value": {
						Description: "A human-readable phone number. It may be in any telephone number format.",
						Type:        types.StringType,
						Required:    true,
					},
				}, tfsdk.ListNestedAttributesOptions{}),
			},
			"suspension_reason": {
				Description: "Has the reason a user account is suspended either by the administrator or by Google at " +
					"the time of suspension. The property is returned only if the suspended property is true.",
				Type:     types.StringType,
				Computed: true,
			},
			"thumbnail_photo_url": {
				Description: "Photo Url of the user.",
				Type:        types.StringType,
				Computed:    true,
			},
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
			"languages": {
				Description: "A list of the user's languages. The maximum allowed data size is 1Kb.",
				Optional:    true,
				Computed:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"custom_language": {
						Description: "Other language. A user can provide their own language name if there is no " +
							"corresponding Google III language code. If this is set, LanguageCode can't be set.",
						Type:     types.StringType,
						Optional: true,
					},
					"language_code": {
						Description: "Language Code. Should be used for storing Google III LanguageCode string " +
							"representation for language. Illegal values cause SchemaException.",
						Type:     types.StringType,
						Optional: true,
						Computed: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							DefaultModifier{
								DefaultValue: types.String{Value: "en"},
							},
						},
						// TODO: (mbang) https://github.com/hashicorp/terraform-plugin-sdk/issues/470
						//ExactlyOneOf: []string{"custom_language", "language_code"},
					},
					"preference": {
						Description: "If present, controls whether the specified languageCode is the user's " +
							"preferred language. Allowed values are `preferred` and `not_preferred`.",
						Type:     types.StringType,
						Optional: true,
						Computed: true,
						PlanModifiers: []tfsdk.AttributePlanModifier{
							DefaultModifier{
								DefaultValue: types.String{Value: "preferred"},
							},
						},
						// TODO: (mbang) https://github.com/hashicorp/terraform-plugin-sdk/issues/470
						//ConflictsWith: []string{"custom_language", "preference"},
					},
				}, tfsdk.ListNestedAttributesOptions{}),
				PlanModifiers: []tfsdk.AttributePlanModifier{
					UserLanguagesModifier{},
				},
			},
			// TODO: (mbang) AtLeastOneOf (https://github.com/hashicorp/terraform-plugin-sdk/issues/470)
			"posix_accounts": {
				Description: "A list of POSIX account information for the user.",
				Optional:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"account_id": {
						Description: "A POSIX account field identifier.",
						Type:        types.StringType,
						Optional:    true,
					},
					"gecos": {
						Description: "The GECOS (user information) for this account.",
						Type:        types.StringType,
						Optional:    true,
					},
					"gid": {
						Description: "The default group ID.",
						Type:        types.StringType,
						Optional:    true,
					},
					"home_directory": {
						Description: "The path to the home directory for this account.",
						Type:        types.StringType,
						Optional:    true,
					},
					"operating_system_type": {
						Description: "The operating system type for this account. " +
							"Acceptable values: `linux`, `unspecified`, `windows`.",
						Type:     types.StringType,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							StringInSliceValidator{
								Options: []string{"linux", "unspecified", "windows"},
							},
						},
					},
					"primary": {
						Description: "If this is user's primary account within the SystemId.",
						Type:        types.BoolType,
						Optional:    true,
					},
					"shell": {
						Description: "The path to the login shell for this account.",
						Type:        types.StringType,
						Optional:    true,
					},
					"system_id": {
						Description: "System identifier for which account Username or Uid apply to.",
						Type:        types.StringType,
						Optional:    true,
					},
					"uid": {
						Description: "The POSIX compliant user ID.",
						Type:        types.StringType,
						Optional:    true,
					},
					"username": {
						Description: "The username of the account.",
						Type:        types.StringType,
						Optional:    true,
					},
				}, tfsdk.ListNestedAttributesOptions{}),
			},
			"creation_time": {
				Description: "The time the user's account was created. The value is in ISO 8601 date and time format. " +
					"The time is the complete date plus hours, minutes, and seconds in the form " +
					"YYYY-MM-DDThh:mm:ssTZD. For example, 2010-04-05T17:30:04+01:00.",
				Type:     types.StringType,
				Computed: true,
			},
			"non_editable_aliases": {
				Description: "asps.list of the user's non-editable alias email addresses. These are typically outside " +
					"the account's primary domain or sub-domain.",
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Computed: true,
			},
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
			"ssh_public_keys": {
				Description: "A list of SSH public keys. The maximum allowed data size is 10Kb.",
				Optional:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"expiration_time_usec": {
						Description: "An expiration time in microseconds since epoch.",
						Type:        types.StringType,
						Optional:    true,
					},
					"fingerprint": {
						Description: "A SHA-256 fingerprint of the SSH public key.",
						Type:        types.StringType,
						Computed:    true,
					},
					"key": {
						Description: "An SSH public key.",
						Type:        types.StringType,
						Required:    true,
					},
				}, tfsdk.ListNestedAttributesOptions{}),
			},
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
			"websites": {
				Description: "A list of the user's websites. The maximum allowed data size is 2Kb.",
				Optional:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"custom_type": {
						Description: "The custom type. Only used if the type is custom.",
						Type:        types.StringType,
						Optional:    true,
					},
					"primary": {
						Description: "If this is user's primary website or not.",
						Type:        types.BoolType,
						Optional:    true,
					},
					"type": {
						Description: "The type or purpose of the website. For example, a website could be labeled " +
							"as home or blog. Alternatively, an entry can have a custom type " +
							"Custom types must have a customType value. " +
							"Acceptable values: `app_install_page`, `blog`, `custom`, `ftp` " +
							", `home`, `home_page`, `other`, `profile`, `reservations`, `resume`, `work`.",
						Type:     types.StringType,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							StringInSliceValidator{
								Options: []string{"app_install_page", "blog", "custom", "ftp",
									"home", "home_page", "other", "profile", "reservations", "resume", "work"},
							},
						},
					},
					"value": {
						Description: "The URL of the website.",
						Type:        types.StringType,
						Required:    true,
					},
				}, tfsdk.ListNestedAttributesOptions{}),
			},
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
			"locations": {
				Description: "A list of the user's locations. The maximum allowed data size is 10Kb.",
				Optional:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"area": {
						Description: "Textual location. This is most useful for display purposes to concisely " +
							"describe the location. For example, Mountain View, CA or Near Seattle.",
						Type:     types.StringType,
						Optional: true,
					},
					"building_id": {
						Description: "Building identifier.",
						Type:        types.StringType,
						Optional:    true,
					},
					"custom_type": {
						Description: "If the location type is custom, this property contains the custom value.",
						Type:        types.StringType,
						Optional:    true,
					},
					"desk_code": {
						Description: "Most specific textual code of individual desk location.",
						Type:        types.StringType,
						Optional:    true,
					},
					"floor_name": {
						Description: "Floor name/number.",
						Type:        types.StringType,
						Optional:    true,
					},
					"floor_section": {
						Description: "Floor section. More specific location within the floor. For example, " +
							"if a floor is divided into sections A, B, and C, this field would identify one " +
							"of those values.",
						Type:     types.StringType,
						Optional: true,
					},
					"type": {
						Description: "The location type." +
							"Acceptable values: `custom`, `default`, `desk`",
						Type:     types.StringType,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							StringInSliceValidator{
								Options: []string{"custom", "default", "desk"},
							},
						},
					},
				}, tfsdk.ListNestedAttributesOptions{}),
			},
			"include_in_global_address_list": {
				Description: "Indicates if the user's profile is visible in the Google Workspace global address list " +
					"when the contact sharing feature is enabled for the domain.",
				Type:     types.BoolType,
				Optional: true,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.Bool{Value: true},
					},
				},
			},
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
			"keywords": {
				Description: "A list of the user's keywords. The maximum allowed data size is 1Kb.",
				Optional:    true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"custom_type": {
						Description: "Custom Type.",
						Type:        types.StringType,
						Optional:    true,
					},
					"type": {
						Description: "Each entry can have a type which indicates standard type of that entry. " +
							"For example, keyword could be of type occupation or outlook. In addition to the " +
							"standard type, an entry can have a custom type and can give it any name. Such types " +
							"should have the CUSTOM value as type and also have a customType value. " +
							"Acceptable values: `custom`, `mission`, `occupation`, `outlook`",
						Type:     types.StringType,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							StringInSliceValidator{
								Options: []string{"custom", "mission", "occupation", "outlook"},
							},
						},
					},
					"value": {
						Description: "Keyword.",
						Type:        types.StringType,
						Required:    true,
					},
				}, tfsdk.ListNestedAttributesOptions{}),
			},
			"deletion_time": {
				Description: "The time the user's account was deleted. The value is in ISO 8601 date and time format " +
					"The time is the complete date plus hours, minutes, and seconds in the form YYYY-MM-DDThh:mm:ssTZD. " +
					"For example 2010-04-05T17:30:04+01:00.",
				Type:     types.StringType,
				Computed: true,
			},
			// "gender" is not included in the GET response, currently, so leaving this out for now
			"thumbnail_photo_etag": {
				Description: "ETag of the user's photo",
				Type:        types.StringType,
				Computed:    true,
			},
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
			"ims": {
				Description: "The user's Instant Messenger (IM) accounts. A user account can have multiple ims " +
					"properties. But, only one of these ims properties can be the primary IM contact. " +
					"The maximum allowed data size is 2Kb.",
				Optional: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"custom_protocol": {
						Description: "If the protocol value is custom_protocol, this property holds the custom " +
							"protocol's string.",
						Type:     types.StringType,
						Optional: true,
					},
					"custom_type": {
						Description: "If the IM type is custom, this property holds the custom type string.",
						Type:        types.StringType,
						Optional:    true,
					},
					"im": {
						Description: "The user's IM network ID.",
						Type:        types.StringType,
						Optional:    true,
					},
					"primary": {
						Description: "If this is the user's primary IM. " +
							"Only one entry in the IM list can have a value of true.",
						Type:     types.BoolType,
						Optional: true,
					},
					"protocol": {
						Description: "An IM protocol identifies the IM network. " +
							"The value can be a custom network or the standard network. " +
							"Acceptable values: `aim`, `custom_protocol`, `gtalk`, `icq`, `jabber`, " +
							"`msn`, `net_meeting`, `qq`, `skype`, `yahoo`.",
						Type:     types.StringType,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							StringInSliceValidator{
								Options: []string{"aim", "custom_protocol", "gtalk", "icq",
									"jabber", "msn", "net_meeting", "qq", "skype", "yahoo"},
							},
						},
					},
					"type": {
						Description: "Acceptable values: `custom`, `home`, `other`, `work`.",
						Type:        types.StringType,
						Required:    true,
						Validators: []tfsdk.AttributeValidator{
							StringInSliceValidator{
								Options: []string{"custom", "home", "other", "work"},
							},
						},
					},
				}, tfsdk.ListNestedAttributesOptions{}),
			},
			"custom_schemas": {
				Description: "Custom fields of the user.",
				//DiffSuppressFunc: diffSuppressCustomSchemas,
				Optional: true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"schema_name": {
						Description: "The name of the schema.",
						Type:        types.StringType,
						Required:    true,
					},
					"schema_values": {
						Description: "JSON encoded map that represents key/value pairs that " +
							"correspond to the given schema. ",
						Type: types.MapType{
							ElemType: types.StringType,
						},
						Required: true,
						//ValidateDiagFunc: validation.ToDiagFunc(
						//	validation.StringIsJSON,
						//	),
					},
				}, tfsdk.ListNestedAttributesOptions{}),
			},
			"is_enrolled_in_2_step_verification": {
				Description: "Is enrolled in 2-step verification.",
				Type:        types.BoolType,
				Computed:    true,
			},
			"is_enforced_in_2_step_verification": {
				Description: "Is 2-step verification enforced.",
				Type:        types.BoolType,
				Computed:    true,
			},
			"archived": {
				Description: "Indicates if user is archived.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultModifier{
						DefaultValue: types.Bool{Value: false},
					},
				},
			},
			"org_unit_path": {
				Description: "The full path of the parent organization associated with the user. " +
					"If the parent organization is the top-level, it is represented as a forward slash (/).",
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"recovery_email": {
				Description: "Recovery email of the user.",
				Type:        types.StringType,
				Optional:    true,
			},
			"recovery_phone": {
				Description: "Recovery phone of the user. The phone number must be in the E.164 format, " +
					"starting with the plus sign (+). Example: +16506661212.",
				Type:     types.StringType,
				Optional: true,
			},
			"id": {
				Computed:            true,
				MarkdownDescription: "User identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

type userResource struct {
	provider provider
}

func (r resourceUserType) NewResource(_ context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	p, diags := convertProviderType(in)

	return userResource{
		provider: p,
	}, diags
}

// Create a new user
func (r userResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from "+
				"another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan model.UserResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// we don't want the planned emails, only what's in config
	// Retrieve values from config
	var config model.UserResourceData
	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userReq := UserPlanToObj(ctx, &r.provider, &config, &plan, &resp.Diagnostics)

	log.Printf("[DEBUG] Creating User %q: %#v", plan.ID.Value, plan.PrimaryEmail.Value)
	usersService := GetUsersService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	userObj, err := usersService.Insert(&userReq).Do()
	if err != nil {
		resp.Diagnostics.AddError("error while trying to create user", err.Error())
		return
	}

	if userObj == nil {
		resp.Diagnostics.AddError(fmt.Sprintf("no user was returned for %s", plan.PrimaryEmail.Value), "object returned was nil")
		return
	}

	userId := userObj.Id

	// The etag changes with each insert, so we want to monitor how many changes we should see
	// when we're checking for eventual consistency
	numInserts := 1

	aliases := plan.Aliases.Elems

	if len(aliases) > 0 {
		aliasesService := GetUserAliasService(&r.provider, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, alias := range aliases {
			aliasName, err := alias.ToTerraformValue(ctx)
			if err != nil {
				return
			}

			aliasObj := directory.Alias{
				Alias: aliasName.String(),
			}

			_, err = aliasesService.Insert(userId, &aliasObj).Do()
			if err != nil {
				return
			}
			numInserts += 1
		}
	}

	// INSERT will respond with the User that will be created, however, it is eventually consistent
	// After INSERT, the etag is updated along with the User (and any aliases),
	// once we get a consistent etag, we can feel confident that our User is also consistent
	cc := consistencyCheck{
		resourceType: "user",
		timeout:      CreateTimeout,
		num404s:      0,
	}
	err = retryTimeDuration(ctx, CreateTimeout, func() error {
		if cc.reachedConsistency(numInserts) {
			return nil
		}

		newUser, retryErr := usersService.Get(userId).IfNoneMatch(cc.lastEtag).Do()
		if googleapi.IsNotModified(retryErr) {
			cc.currConsistent += 1
		} else if retryErr != nil {
			return cc.is404(retryErr)
		} else {
			cc.handleNewEtag(newUser.Etag)
		}

		return fmt.Errorf("timed out while waiting for %s to be inserted", cc.resourceType)
	})
	if err != nil {
		return
	}

	if plan.IsAdmin.Value {
		makeAdminObj := directory.UserMakeAdmin{
			Status:          plan.IsAdmin.Value,
			ForceSendFields: []string{"Status"},
		}

		cc = consistencyCheck{
			resourceType: "user",
			timeout:      CreateTimeout,
			num404s:      0,
		}
		err = usersService.MakeAdmin(userId, &makeAdminObj).Do()
		if err != nil {
			resp.Diagnostics.AddError("error while updating user to admin", err.Error())
			return
		}

		numInserts = 1
		err = retryTimeDuration(ctx, CreateTimeout, func() error {
			if cc.reachedConsistency(numInserts) {
				return nil
			}

			newUser, retryErr := usersService.Get(userId).IfNoneMatch(cc.lastEtag).Do()
			if googleapi.IsNotModified(retryErr) {
				cc.currConsistent += 1
			} else if retryErr != nil {
				return cc.is404(retryErr)
			} else {
				cc.handleNewEtag(newUser.Etag)
			}

			return fmt.Errorf("timed out while waiting for %s to be updated to admin", cc.resourceType)
		})
		if err != nil {
			return
		}
	}

	plan.ID.Value = userId
	user := GetUserData(ctx, &r.provider, plan, &resp.Diagnostics)

	diags = resp.State.Set(ctx, user)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Finished creating User %s: %s", user.ID.Value, user.PrimaryEmail.Value)
}

// Read user information
func (r userResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state model.UserResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := GetUserData(ctx, &r.provider, state, &resp.Diagnostics)
	if user.ID.Null {
		resp.State.RemoveResource(ctx)
		log.Printf("[DEBUG] Removed User from state because it was not found %s", state.ID.Value)
		return
	}

	diags = resp.State.Set(ctx, user)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	log.Printf("[DEBUG] Finished getting User %s: %s", state.ID.Value, user.PrimaryEmail.Value)
}

// Update user resource
func (r userResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Retrieve values from config
	var config model.UserResourceData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from plan
	var plan model.UserResourceData
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state model.UserResourceData
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Updating User %q: %#v", plan.ID.Value, plan.PrimaryEmail.Value)
	usersService := GetUsersService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	userReq := UserPlanToObj(ctx, &r.provider, &config, &plan, &resp.Diagnostics)

	numInserts := 0
	if !reflect.DeepEqual(plan.Aliases, state.Aliases) {
		var stateAliases []string
		for _, sa := range state.Aliases.Elems {
			aliasName, err := sa.ToTerraformValue(ctx)
			if err != nil {
				return
			}
			stateAliases = append(stateAliases, aliasName.String())
		}

		var planAliases []string
		for _, pa := range plan.Aliases.Elems {
			aliasName, err := pa.ToTerraformValue(ctx)
			if err != nil {
				return
			}
			planAliases = append(planAliases, aliasName.String())
		}

		aliasesService := GetUserAliasService(&r.provider, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		// Remove old aliases that aren't in the new aliases list
		for _, alias := range stateAliases {
			if stringInSlice(planAliases, alias) {
				continue
			}

			err := aliasesService.Delete(state.ID.Value, alias).Do()
			if err != nil {
				resp.Diagnostics.AddError(fmt.Sprintf("error deleting alias (%s) from user (%s)", alias, state.PrimaryEmail.Value), err.Error())
				return
			}
		}

		// Insert all new aliases that weren't previously in state
		for _, alias := range planAliases {
			if stringInSlice(stateAliases, alias) {
				continue
			}

			aliasObj := directory.Alias{
				Alias: alias,
			}

			_, err := aliasesService.Insert(state.ID.Value, &aliasObj).Do()
			if err != nil {
				resp.Diagnostics.AddError(fmt.Sprintf("error inserting alias (%s) into user (%s)", alias, state.PrimaryEmail.Value), err.Error())
				return
			}
			numInserts += 1
		}
	}

	if plan.IsAdmin.Value != state.IsAdmin.Value {
		makeAdminObj := directory.UserMakeAdmin{
			Status:          plan.IsAdmin.Value,
			ForceSendFields: []string{"Status"},
		}

		err := usersService.MakeAdmin(state.ID.Value, &makeAdminObj).Do()
		if err != nil {
			resp.Diagnostics.AddError("error while updating user to admin", err.Error())
			return
		}
		numInserts += 1
	}

	_, err := usersService.Update(state.ID.Value, &userReq).Do()
	if err != nil {
		resp.Diagnostics.AddError("error while trying to update user", err.Error())
		return
	}
	numInserts += 1

	// UPDATE will respond with the User that will be created, however, it is eventually consistent
	// After UPDATE, the etag is updated along with the User (and any aliases),
	// once we get a consistent etag, we can feel confident that our User is also consistent
	cc := consistencyCheck{
		resourceType: "user",
		timeout:      UpdateTimeout,
	}
	err = retryTimeDuration(ctx, UpdateTimeout, func() error {
		var retryErr error

		if cc.reachedConsistency(numInserts) {
			return nil
		}

		newUser, retryErr := usersService.Get(state.ID.Value).IfNoneMatch(cc.lastEtag).Do()
		if googleapi.IsNotModified(retryErr) {
			cc.currConsistent += 1
		} else if retryErr != nil {
			return fmt.Errorf("unexpected error during retries of %s: %s", cc.resourceType, retryErr)
		} else {
			cc.handleNewEtag(newUser.Etag)
		}

		return fmt.Errorf("timed out while waiting for %s to be updated", cc.resourceType)
	})
	if err != nil {
		resp.Diagnostics.AddError("error while trying to update user", err.Error())
		return
	}

	user := GetUserData(ctx, &r.provider, state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, user)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Finished creating User %q: %#v", state.ID.Value, plan.PrimaryEmail.Value)
}

// Delete user
func (r userResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state model.UserResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Deleting User %q: %#v", state.ID.Value, state.ID.Value)
	usersService := GetUsersService(&r.provider, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := usersService.Delete(state.PrimaryEmail.Value).Do()
	if err != nil {
		state.ID = types.String{Value: handleNotFoundError(err, state.ID.Value, &resp.Diagnostics)}
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.State.RemoveResource(ctx)
	log.Printf("[DEBUG] Finished deleting User %s: %s", state.ID.Value, state.PrimaryEmail.Value)
}

// ImportState user
func (r userResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

//func diffSuppressCustomSchemas(_, _, _ string, d *schema.ResourceData) bool {
//	old, new := d.GetChange("custom_schemas")
//	customSchemasOld := old.([]interface{})
//	customSchemasNew := new.([]interface{})
//
//	// transform the blocks
//	//
//	// custom_schemas {
//	// 	schema_name = "a"
//	//
//	// 	schema_values = {
//	// 	  "bar" = jsonencode("Bar")
//	// 	}
//	// }
//	//
//	// custom_schemas {
//	// 	schema_name = "b"
//	//
//	// 	schema_values = {
//	// 	  "baz" = jsonencode("Baz")
//	// 	}
//	// }
//	//
//	// into a 2 dimentional map[string]map[string]string
//	//
//	// {
//	// 	"a": {
//	// 		"bar": "Bar",
//	// 	},
//	// 	"b": {
//	// 		"baz": "Baz",
//	// 	},
//	// }
//	//
//	// and use reflect.DeepEqual to compare
//
//	oldMap := transformCustomSchemasTo2DMap(customSchemasOld)
//	newMap := transformCustomSchemasTo2DMap(customSchemasNew)
//
//	return reflect.DeepEqual(oldMap, newMap)
//}
//
//func transformCustomSchemasTo2DMap(customSchemas []interface{}) map[string]map[string]string {
//	result := make(map[string]map[string]string)
//	for _, schema := range customSchemas {
//		s := schema.(map[string]interface{})
//		schemaValues := make(map[string]string)
//		for k, v := range s["schema_values"].(map[string]interface{}) {
//			// ensure if field is list that it is sorted for comparison
//			// google stores unordered multi-value fields
//			var list []interface{}
//			if err := json.Unmarshal([]byte(v.(string)), &list); err == nil {
//				sorted := sortListOfInterfaces(list)
//				encoded, err := json.Marshal(sorted)
//				if err != nil {
//					panic(err)
//				}
//				schemaValues[k] = string(encoded)
//			} else {
//				schemaValues[k] = v.(string)
//			}
//		}
//		result[s["schema_name"].(string)] = schemaValues
//	}
//	return result
//}

// Helper functions

// Custom Schemas

func validateCustomSchemas(ctx context.Context, newSchemas []attr.Value, p *provider, diags *diag.Diagnostics) {
	schemaService := GetSchemasService(p, diags)
	if diags.HasError() {
		return
	}

	// Validate config against schemas
	for _, cs := range newSchemas {
		customSchema := userCustomSchemaData{}
		d := cs.(types.Object).As(ctx, &customSchema, types.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return
		}

		schemaName := customSchema.SchemaName.Value

		schemaDef, err := schemaService.Get(p.customer, schemaName).Do()
		if err != nil {
			diags.AddError(fmt.Sprintf("error getting custom schema %s", schemaName), err.Error())
			return
		}

		if schemaDef == nil {
			diags.AddError(fmt.Sprintf("error getting custom schema %s", schemaName), "object was returned nil")
			return
		}

		schemaFieldMap := map[string]*directory.SchemaFieldSpec{}
		for _, schemaField := range schemaDef.Fields {
			schemaFieldMap[schemaField.FieldName] = schemaField
		}

		customSchemaDef := customSchema.SchemaValues

		for csKey, csJsonVal := range customSchemaDef.Elems {
			if _, ok := schemaFieldMap[csKey]; !ok {
				diags.AddError("field name is not found in schema", fmt.Sprintf("field name (%s) is not found in this schema definition (%s)", csKey, schemaName))
				return
			}

			var csVal interface{}
			err := json.Unmarshal([]byte(csJsonVal.(types.String).Value), &csVal)
			if err != nil {
				diags.AddError("error while unmarshalling", err.Error())
				return
			}

			if schemaFieldMap[csKey].MultiValued {
				if reflect.ValueOf(csVal).Kind() != reflect.Slice {
					diags.AddError("field value should be a list", fmt.Sprintf("field %s is multi-values and should be a list (%+v)", csKey, csVal))
					return
				}

				if len(csVal.([]interface{})) > 0 {
					csVal = csVal.([]interface{})[0]
				}
			}

			validType := validateSchemaFieldValueType(schemaFieldMap[csKey].FieldType, csVal)
			if !validType {
				diags.AddError("value in custom schema is of incorrect type", fmt.Sprintf("value provided for %s is of incorrect type (expected type: %s)", csKey, schemaFieldMap[csKey].FieldType))
				return
			}
		}
	}

	return
}

// This will take a value and validate whether the type is correct
func validateSchemaFieldValueType(fieldType string, fieldValue interface{}) bool {
	valid := false

	switch fieldType {
	case "BOOL":
		valid = reflect.ValueOf(fieldValue).Kind() == reflect.Bool
	case "DATE":
		// ISO 8601 format
		_, err := time.Parse("2006-01-02", fieldValue.(string))
		if err == nil {
			valid = true
		}
	case "DOUBLE":
		valid = reflect.ValueOf(fieldValue).Kind() == reflect.Float64
	case "EMAIL":
		_, err := mail.ParseAddress(fieldValue.(string))
		if err == nil {
			valid = true
		}
	case "INT64":
		// this is unmarshalled as a float, check that it's an int
		if reflect.ValueOf(fieldValue).Kind() == reflect.Float64 &&
			fieldValue == float64(int(fieldValue.(float64))) {
			valid = true
		}
	case "PHONE":
		fallthrough
	case "STRING":
		fallthrough
	default:
		valid = reflect.ValueOf(fieldValue).Kind() == reflect.String
	}

	return valid
}

func expandCustomSchemaValues(ctx context.Context, customSchemas []attr.Value, diags *diag.Diagnostics) map[string]googleapi.RawMessage {
	result := map[string]googleapi.RawMessage{}

	for _, cs := range customSchemas {
		customSchema := userCustomSchemaData{}
		d := cs.(types.Object).As(ctx, &customSchema, types.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return result
		}
		schemaName := customSchema.SchemaName.Value
		schemaValues := customSchema.SchemaValues

		customSchemaObj := map[string]interface{}{}
		for k, v := range schemaValues.Elems {
			var csVal interface{}
			err := json.Unmarshal([]byte(v.(types.String).Value), &csVal)
			if err != nil {
				diags.AddError("error unmarshalling custom schemas", err.Error())
				return nil
			}

			if reflect.ValueOf(csVal).Kind() == reflect.Slice {
				newSlice := []map[string]interface{}{}
				for _, nested := range csVal.([]interface{}) {
					newSlice = append(newSlice, map[string]interface{}{
						"type":  "work",
						"value": nested,
					})
				}

				customSchemaObj[k] = newSlice
			} else {
				customSchemaObj[k] = csVal
			}
		}
		// create the json object and assign to the schema
		schemaValuesJson, err := json.Marshal(customSchemaObj)
		if err != nil {
			diags.AddError("error marshalling custom schemas", err.Error())
			return nil
		}

		result[schemaName] = schemaValuesJson
	}

	return result
}

// The API returns numeric values as strings. This will convert it to the appropriate type
func convertSchemaFieldValueType(fieldType string, fieldValue interface{}) (interface{}, error) {
	// If it's not of type string, then we'll assume it's the right type
	if reflect.ValueOf(fieldValue).Kind() != reflect.String {
		return fieldValue, nil
	}

	var err error
	var value interface{}

	switch fieldType {
	case "BOOL":
		value, err = strconv.ParseBool(fieldValue.(string))
	case "DOUBLE":
		value, err = strconv.ParseFloat(fieldValue.(string), 64)
	case "INT64":
		value, err = strconv.ParseInt(fieldValue.(string), 10, 64)
	case "DATE":
		// The string stays the same
		fallthrough
	case "EMAIL":
		fallthrough
	case "PHONE":
		fallthrough
	case "STRING":
		fallthrough
	default:
		value = fieldValue
	}

	return value, err
}

func flattenCustomSchemas(prov *provider, schemaAttrObj interface{}) ([]attr.Value, diag.Diagnostics) {
	var customSchemas []attr.Value
	var diags diag.Diagnostics

	schemaService := GetSchemasService(prov, &diags)
	if diags.HasError() {
		return nil, diags
	}

	for schemaName, sv := range schemaAttrObj.(map[string]googleapi.RawMessage) {
		schemaDef, err := schemaService.Get(prov.customer, schemaName).Do()
		if err != nil {
			diags.AddError(fmt.Sprintf("error reading schema (%s)", schemaName), err.Error())
			return nil, diags
		}

		schemaFieldMap := map[string]*directory.SchemaFieldSpec{}
		for _, schemaField := range schemaDef.Fields {
			schemaFieldMap[schemaField.FieldName] = schemaField
		}

		var schemaValuesObj map[string]interface{}

		err = json.Unmarshal(sv, &schemaValuesObj)
		if err != nil {
			diags.AddError(fmt.Sprintf("error unmarshalling schema values (%s): %+v", schemaName, sv), err.Error())
			return nil, diags
		}

		schemaValues := types.Map{}
		for k, v := range schemaValuesObj {
			if _, ok := schemaFieldMap[k]; !ok {
				diags.AddError(fmt.Sprintf("field name (%s) is not found in this schema definition (%s)", k, schemaName), "correct field name is required")
				return nil, diags
			}

			if schemaFieldMap[k].MultiValued {
				vals := []interface{}{}
				for _, item := range v.([]interface{}) {
					val, err := convertSchemaFieldValueType(schemaFieldMap[k].FieldType, item.(map[string]interface{})["value"])
					if err != nil {
						diags.AddError(fmt.Sprintf("could not properly convert field value type for %+v: %+v", schemaFieldMap[k].FieldType, item.(map[string]interface{})["value"]), err.Error())
						return nil, diags
					}
					vals = append(vals, val)
				}
				jsonVals, err := json.Marshal(vals)
				if err != nil {
					diags.AddError(fmt.Sprintf("error marshalling schema values %+v", vals), err.Error())
					return nil, diags
				}
				schemaValues = types.Map{
					ElemType: types.StringType,
					Elems: map[string]attr.Value{
						k: types.String{Value: string(jsonVals)},
					},
				}
			} else {
				val, err := convertSchemaFieldValueType(schemaFieldMap[k].FieldType, v)
				if err != nil {
					diags.AddError(fmt.Sprintf("could not properly convert field value type for %+v: %+v", schemaFieldMap[k].FieldType, v), err.Error())
					return nil, diags
				}

				jsonVal, err := json.Marshal(val)
				if err != nil {
					diags.AddError(fmt.Sprintf("error marshalling schema values %+v", val), err.Error())
					return nil, diags
				}
				schemaValues = types.Map{
					ElemType: types.StringType,
					Elems: map[string]attr.Value{
						k: types.String{Value: string(jsonVal)},
					},
				}
			}
		}

		customSchemas = append(customSchemas, types.Object{
			AttrTypes: map[string]attr.Type{
				"schema_name": types.StringType,
				"schema_values": types.MapType{
					ElemType: types.StringType,
				},
			},
			Attrs: map[string]attr.Value{
				"schema_name":   types.String{Value: schemaName},
				"schema_values": schemaValues,
			},
		})
	}

	return customSchemas, nil
}
