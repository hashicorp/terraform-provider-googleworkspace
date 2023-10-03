// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/mail"
	"reflect"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/sethvargo/go-password/password"

	directory "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
)

func diffSuppressEmails(k, old, new string, d *schema.ResourceData) bool {
	stateEmails, configEmails := d.GetChange("emails")

	// User aliases and other alternate emails (added by Google and denoted by no `type`),
	// are auto-added to the email list, even if it's not configured that way.
	// Only show a diff if the other emails differ.
	subsetEmails := []interface{}{}

	aliases := listOfInterfacestoStrings(d.Get("aliases").([]interface{}))

	for _, se := range stateEmails.([]interface{}) {
		email := se.(map[string]interface{})
		emailAddress := email["address"].(string)
		emailType := email["type"].(string)

		if emailType == "" || stringInSlice(aliases, emailAddress) {
			continue
		}

		subsetEmails = append(subsetEmails, se)
	}

	return reflect.DeepEqual(subsetEmails, configEmails.([]interface{}))
}

func diffSuppressCustomSchemas(_, _, _ string, d *schema.ResourceData) bool {
	old, new := d.GetChange("custom_schemas")
	customSchemasOld := old.([]interface{})
	customSchemasNew := new.([]interface{})

	// transform the blocks
	//
	// custom_schemas {
	// 	schema_name = "a"
	//
	// 	schema_values = {
	// 	  "bar" = jsonencode("Bar")
	// 	}
	// }
	//
	// custom_schemas {
	// 	schema_name = "b"
	//
	// 	schema_values = {
	// 	  "baz" = jsonencode("Baz")
	// 	}
	// }
	//
	// into a 2 dimentional map[string]map[string]string
	//
	// {
	// 	"a": {
	// 		"bar": "Bar",
	// 	},
	// 	"b": {
	// 		"baz": "Baz",
	// 	},
	// }
	//
	// and use reflect.DeepEqual to compare

	oldMap := transformCustomSchemasTo2DMap(customSchemasOld)
	newMap := transformCustomSchemasTo2DMap(customSchemasNew)

	return reflect.DeepEqual(oldMap, newMap)
}

func transformCustomSchemasTo2DMap(customSchemas []interface{}) map[string]map[string]string {
	result := make(map[string]map[string]string)
	for _, schema := range customSchemas {
		s := schema.(map[string]interface{})
		schemaValues := make(map[string]string)
		for k, v := range s["schema_values"].(map[string]interface{}) {
			// ensure if field is list that it is sorted for comparison
			// google stores unordered multi-value fields
			var list []interface{}
			if err := json.Unmarshal([]byte(v.(string)), &list); err == nil {
				sorted := sortListOfInterfaces(list)
				encoded, err := json.Marshal(sorted)
				if err != nil {
					panic(err)
				}
				schemaValues[k] = string(encoded)
			} else {
				schemaValues[k] = v.(string)
			}
		}
		result[s["schema_name"].(string)] = schemaValues
	}
	return result
}

func resourceUser() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "User resource manages Google Workspace Users. User resides " +
			"under the `https://www.googleapis.com/auth/admin.directory.user` client scope.",

		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
					"As the API does not return the value of password, this field is write-only, and the value stored " +
					"in the state will be what is provided in the configuration. If the field is not set on create a random password will be generated " +
					"be empty on import.",
				Type:             schema.TypeString,
				Optional:         true,
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
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
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
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
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
							Required: true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"custom", "home", "other", "work"}, false),
							),
						},
					},
				},
			},
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
			"external_ids": {
				Description: "A list of external IDs for the user, such as an employee or network ID. " +
					"The maximum allowed data size is 2Kb.",
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_type": {
							Description: "If the external ID type is custom, this property contains the custom value and " +
								"must be set.",
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Description: "The type of external ID. If set to custom, customType must also be set. " +
								"Acceptable values: `account`, `custom`, `customer`, `login_id`, `network`, `organization`.",
							Type:     schema.TypeString,
							Required: true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"account", "custom", "customer", "login_id",
									"network", "organization"}, false),
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
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
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
							Required: true,
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
				Optional:    true,
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
			// And add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
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
							Required: true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"custom", "home", "other", "work"}, false),
							),
						},
					},
				},
			},
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
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
							Required: true,
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
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
			"phones": {
				Description: "A list of the user's phone numbers. The maximum allowed data size is 1Kb.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_type": {
							Description: "If the phone number type is custom, this property contains the custom value " +
								"and must be set.",
							Type:     schema.TypeString,
							Optional: true,
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
							Required: true,
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
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
			"languages": {
				Description: "A list of the user's languages. The maximum allowed data size is 1Kb.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_language": {
							Description: "Other language. A user can provide their own language name if there is no " +
								"corresponding Google III language code. If this is set, LanguageCode can't be set.",
							Type:     schema.TypeString,
							Optional: true,
							// TODO: (mbang) https://github.com/hashicorp/terraform-plugin-sdk/issues/470
							//ExactlyOneOf: []string{"custom_language", "language_code"},
							//ConflictsWith: []string{"custom_language", "preference"},
						},
						"language_code": {
							Description: "Language Code. Should be used for storing Google III LanguageCode string " +
								"representation for language. Illegal values cause SchemaException.",
							Type:     schema.TypeString,
							Optional: true,
							Default:  "en",
							// TODO: (mbang) https://github.com/hashicorp/terraform-plugin-sdk/issues/470
							//ExactlyOneOf: []string{"custom_language", "language_code"},
						},
						"preference": {
							Description: "If present, controls whether the specified languageCode is the user's " +
								"preferred language. Allowed values are `preferred` and `not_preferred`.",
							Type:     schema.TypeString,
							Optional: true,
							Default:  "preferred",
							// TODO: (mbang) https://github.com/hashicorp/terraform-plugin-sdk/issues/470
							//ConflictsWith: []string{"custom_language", "preference"},
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
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
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
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
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
							Required: true,
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
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
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
							Required: true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"custom", "default", "desk"},
									false),
							),
						},
					},
				},
			},
			"include_in_global_address_list": {
				Description: "Indicates if the user's profile is visible in the Google Workspace global address list " +
					"when the contact sharing feature is enabled for the domain.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
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
							Required: true,
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
			// TODO: (mbang) Add ValidateDiagFunc for max size when it's allowed on lists
			// (https://github.com/hashicorp/terraform-plugin-sdk/issues/156)
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
							Required: true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"aim", "custom_protocol", "gtalk", "icq",
									"jabber", "msn", "net_meeting", "qq", "skype", "yahoo"}, false),
							),
						},
						"type": {
							Description: "Acceptable values: `custom`, `home`, `other`, `work`.",
							Type:        schema.TypeString,
							Required:    true,
							ValidateDiagFunc: validation.ToDiagFunc(
								validation.StringInSlice([]string{"custom", "home", "other", "work"}, false),
							),
						},
					},
				},
			},
			"custom_schemas": {
				Description:      "Custom fields of the user.",
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: diffSuppressCustomSchemas,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"schema_name": {
							Description: "The name of the schema.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"schema_values": {
							Description: "JSON encoded map that represents key/value pairs that " +
								"correspond to the given schema. ",
							Type:     schema.TypeMap,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateDiagFunc: validation.ToDiagFunc(
									validation.StringIsJSON,
								),
							},
						},
					},
				},
			},
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
	var generated_password string
	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)
	generated_password = ""
	if d.Get("password").(string) == "" {
		// generate password
		generated_password, _ = password.Generate(rand.Intn(64)+8, 4, 4, false, true)
		log.Printf("[DEBUG] Auto Generating password for User %q", d.Id())
	} else {
		generated_password = d.Get("password").(string)
	}

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
		Password:                   generated_password,
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

	if len(d.Get("custom_schemas").([]interface{})) > 0 {
		diags = validateCustomSchemas(d, client)
		if diags.HasError() {
			return diags
		}

		customSchemas, diags := expandCustomSchemaValues(d.Get("custom_schemas").([]interface{}))
		if diags.HasError() {
			return diags
		}

		userObj.CustomSchemas = customSchemas
	}

	user, err := usersService.Insert(&userObj).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(user.Id)

	// INSERT will respond with the User that will be created, however, it is eventually consistent
	// After INSERT, the etag is updated along with the User (and any aliases),
	// once we get a consistent etag, we can feel confident that our User is also consistent
	cc := consistencyCheck{
		resourceType: "user",
		timeout:      d.Timeout(schema.TimeoutCreate),
	}
	err = retryTimeDuration(ctx, d.Timeout(schema.TimeoutCreate), func() error {
		var retryErr error

		if cc.reachedConsistency(1) {
			return nil
		}

		newUser, retryErr := usersService.Get(d.Id()).IfNoneMatch(cc.lastEtag).Do()
		if googleapi.IsNotModified(retryErr) {
			cc.currConsistent += 1
		} else if isNotFound(retryErr) {
			// user was not found yet therefore setting currConsistent back to null value
			cc.currConsistent = 0
		} else if retryErr != nil {
			return fmt.Errorf("unexpected error during retries of %s: %s", cc.resourceType, retryErr)
		} else {
			cc.handleNewEtag(newUser.Etag)
		}

		return fmt.Errorf("timed out while waiting for %s to be inserted", cc.resourceType)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	diags = resourceUserUpdate(ctx, d, meta)
	if diags.HasError() {
		return diags
	}

	log.Printf("[DEBUG] Finished creating User %q: %#v", d.Id(), primaryEmail)
	return resourceUserRead(ctx, d, meta)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// use the meta value to retrieve your client from the provider configure method
	client := meta.(*apiClient)

	primaryEmail := d.Get("primary_email").(string)
	log.Printf("[DEBUG] Getting User %q: %#v", d.Id(), primaryEmail)

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
		return handleNotFoundError(err, d, primaryEmail)
	}

	if user == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("No user was returned for %s.", d.Get("primary_email").(string)),
		})

		return diags
	}

	customSchemas := []map[string]interface{}{}
	if len(user.CustomSchemas) > 0 {
		customSchemas, diags = flattenCustomSchemas(user.CustomSchemas, client)
		if diags.HasError() {
			return diags
		}
	}

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
	d.Set("emails", flattenInterfaceObjects(user.Emails))
	d.Set("external_ids", flattenInterfaceObjects(user.ExternalIds))
	d.Set("relations", flattenInterfaceObjects(user.Relations))
	d.Set("etag", user.Etag)
	d.Set("aliases", user.Aliases)
	d.Set("is_mailbox_setup", user.IsMailboxSetup)
	d.Set("customer_id", user.CustomerId)
	d.Set("addresses", flattenInterfaceObjects(user.Addresses))
	d.Set("organizations", flattenInterfaceObjects(user.Organizations))
	d.Set("last_login_time", user.LastLoginTime)
	d.Set("phones", flattenInterfaceObjects(user.Phones))
	d.Set("suspension_reason", user.SuspensionReason)
	d.Set("thumbnail_photo_url", user.ThumbnailPhotoUrl)
	d.Set("languages", flattenInterfaceObjects(user.Languages))
	d.Set("posix_accounts", flattenInterfaceObjects(user.PosixAccounts))
	d.Set("creation_time", user.CreationTime)
	d.Set("non_editable_aliases", user.NonEditableAliases)
	d.Set("ssh_public_keys", flattenInterfaceObjects(user.SshPublicKeys))
	d.Set("websites", flattenInterfaceObjects(user.Websites))
	d.Set("locations", flattenInterfaceObjects(user.Locations))
	d.Set("include_in_global_address_list", user.IncludeInGlobalAddressList)
	d.Set("keywords", flattenInterfaceObjects(user.Keywords))
	d.Set("deletion_time", user.DeletionTime)
	d.Set("thumbnail_photo_etag", user.ThumbnailPhotoEtag)
	d.Set("ims", flattenInterfaceObjects(user.Ims))
	d.Set("custom_schemas", customSchemas)
	d.Set("is_enrolled_in_2_step_verification", user.IsEnrolledIn2Sv)
	d.Set("is_enforced_in_2_step_verification", user.IsEnforcedIn2Sv)
	d.Set("archived", user.Archived)
	d.Set("org_unit_path", user.OrgUnitPath)
	d.Set("recovery_email", user.RecoveryEmail)
	d.Set("recovery_phone", user.RecoveryPhone)

	d.SetId(user.Id)
	log.Printf("[DEBUG] Finished getting User %q: %#v", d.Id(), primaryEmail)

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
	forceSendFields := []string{}

	// Strings

	if d.HasChange("primary_email") {
		userObj.PrimaryEmail = primaryEmail
	}
	if d.Get("password").(string) != "" {
		if d.HasChange("password") {
			userObj.Password = d.Get("password").(string)
		}
	}

	if d.HasChange("hash_function") {
		userObj.HashFunction = d.Get("hash_function").(string)

	}

	if d.HasChange("org_unit_path") {
		userObj.OrgUnitPath = d.Get("org_unit_path").(string)
	}

	if d.HasChange("recovery_email") {
		userObj.RecoveryEmail = d.Get("recovery_email").(string)

		if userObj.RecoveryEmail == "" {
			forceSendFields = append(forceSendFields, "RecoveryEmail")
		}
	}

	if d.HasChange("recovery_phone") {
		userObj.RecoveryPhone = d.Get("recovery_phone").(string)

		if userObj.RecoveryPhone == "" {
			forceSendFields = append(forceSendFields, "RecoveryPhone")
		}
	}

	// Booleans

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

	if d.HasChange("custom_schemas") {
		if len(d.Get("custom_schemas").([]interface{})) > 0 {
			diags = validateCustomSchemas(d, client)
			if diags.HasError() {
				return diags
			}

			customSchemas, diags := expandCustomSchemaValues(d.Get("custom_schemas").([]interface{}))
			if diags.HasError() {
				return diags
			}

			userObj.CustomSchemas = customSchemas
		}
	}

	numInserts := 0
	if d.HasChange("aliases") {
		old, new := d.GetChange("aliases")
		oldAliases := listOfInterfacestoStrings(old.([]interface{}))
		newAliases := listOfInterfacestoStrings(new.([]interface{}))

		aliasesService, diags := GetUserAliasService(usersService)
		if diags.HasError() {
			return diags
		}

		// Remove old aliases that aren't in the new aliases list
		for _, alias := range oldAliases {
			if stringInSlice(newAliases, alias) {
				continue
			}

			err := aliasesService.Delete(d.Id(), alias).Do()
			if err != nil {
				return diag.FromErr(err)
			}
			numInserts += 1
		}

		// Insert all new aliases that weren't previously in state
		for _, alias := range newAliases {
			if stringInSlice(oldAliases, alias) {
				continue
			}

			aliasObj := directory.Alias{
				Alias: alias,
			}

			_, err := aliasesService.Insert(d.Id(), &aliasObj).Do()
			if err != nil {
				return diag.FromErr(err)
			}
			numInserts += 1
		}
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
		numInserts += 1
	}

	if &userObj != new(directory.User) {
		_, err := usersService.Update(d.Id(), &userObj).Do()
		if err != nil {
			return diag.FromErr(err)
		}
		numInserts += 1
	}

	// UPDATE will respond with the updated User, however, it is eventually consistent
	// After UPDATE, the etag is updated along with the User (and any aliases),
	// once we get a consistent etag, we can feel confident that our User is also consistent
	cc := consistencyCheck{
		resourceType: "user",
		timeout:      d.Timeout(schema.TimeoutUpdate),
	}
	err := retryTimeDuration(ctx, d.Timeout(schema.TimeoutUpdate), func() error {
		var retryErr error

		if cc.reachedConsistency(numInserts) {
			return nil
		}

		newUser, retryErr := usersService.Get(d.Id()).IfNoneMatch(cc.lastEtag).Do()
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
		return diag.FromErr(err)
	}

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
		return handleNotFoundError(err, d, primaryEmail)
	}

	log.Printf("[DEBUG] Finished deleting User %q: %#v", d.Id(), primaryEmail)

	return diags
}

// Expand functions

func expandName(v interface{}) *directory.UserName {
	name := v.([]interface{})

	if len(name) == 0 {
		return nil
	}

	nameObj := directory.UserName{
		FamilyName: name[0].(map[string]interface{})["family_name"].(string),
		GivenName:  name[0].(map[string]interface{})["given_name"].(string),
	}
	return &nameObj
}

// Flatten functions

func flattenName(nameObj *directory.UserName) interface{} {
	name := []map[string]interface{}{}

	if nameObj != nil {
		name = append(name, map[string]interface{}{
			"family_name": nameObj.FamilyName,
			"full_name":   nameObj.FullName,
			"given_name":  nameObj.GivenName,
		})
	}

	return name
}

// Helper functions

// Custom Schemas

func validateCustomSchemas(d *schema.ResourceData, client *apiClient) diag.Diagnostics {
	var diags diag.Diagnostics

	new := d.Get("custom_schemas")

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return diags
	}

	schemaService, diags := GetSchemasService(directoryService)
	if diags.HasError() {
		return diags
	}

	// Validate config against schemas
	for _, customSchema := range new.([]interface{}) {
		schemaName := customSchema.(map[string]interface{})["schema_name"].(string)

		schemaDef, err := schemaService.Get(client.Customer, schemaName).Do()
		if err != nil {
			return diag.FromErr(err)
		}

		if schemaDef == nil {
			return append(diags, diag.Diagnostic{
				Summary:  fmt.Sprintf("schema definition (%s) is empty", schemaName),
				Severity: diag.Error,
			})
		}

		schemaFieldMap := map[string]*directory.SchemaFieldSpec{}
		for _, schemaField := range schemaDef.Fields {
			schemaFieldMap[schemaField.FieldName] = schemaField
		}

		customSchemaDef := customSchema.(map[string]interface{})["schema_values"].(map[string]interface{})

		for csKey, csJsonVal := range customSchemaDef {
			if _, ok := schemaFieldMap[csKey]; !ok {
				return append(diags, diag.Diagnostic{
					Summary:  fmt.Sprintf("field name (%s) is not found in this schema definition (%s)", csKey, schemaName),
					Severity: diag.Error,
				})
			}

			var csVal interface{}
			err := json.Unmarshal([]byte(csJsonVal.(string)), &csVal)
			if err != nil {
				return diag.FromErr(err)
			}

			if schemaFieldMap[csKey].MultiValued {
				if reflect.ValueOf(csVal).Kind() != reflect.Slice {
					return append(diags, diag.Diagnostic{
						Summary:  fmt.Sprintf("field %s is multi-values and should be a list (%+v)", csKey, csVal),
						Severity: diag.Error,
					})
				}

				if len(csVal.([]interface{})) > 0 {
					csVal = csVal.([]interface{})[0]
				}
			}

			validType := validateFieldValueType(schemaFieldMap[csKey].FieldType, csVal)
			if !validType {
				return append(diags, diag.Diagnostic{
					Summary:  fmt.Sprintf("value provided for %s is of incorrect type (expected type: %s)", csKey, schemaFieldMap[csKey].FieldType),
					Severity: diag.Error,
				})
			}
		}
	}

	return nil
}

// This will take a value and validate whether the type is correct
func validateFieldValueType(fieldType string, fieldValue interface{}) bool {
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

// The API returns numeric values as strings. This will convert it to the appropriate type
func convertFieldValueType(fieldType string, fieldValue interface{}) (interface{}, error) {
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

func expandCustomSchemaValues(customSchemas []interface{}) (map[string]googleapi.RawMessage, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := map[string]googleapi.RawMessage{}

	for _, cs := range customSchemas {
		customSchema := cs.(map[string]interface{})

		schemaName := customSchema["schema_name"].(string)
		schemaValues := customSchema["schema_values"].(map[string]interface{})

		customSchemaObj := map[string]interface{}{}
		for k, v := range schemaValues {
			var csVal interface{}
			err := json.Unmarshal([]byte(v.(string)), &csVal)
			if err != nil {
				return nil, diag.FromErr(err)
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
			return nil, diag.FromErr(err)
		}

		result[schemaName] = schemaValuesJson
	}

	return result, diags
}

func flattenCustomSchemas(schemaAttrObj interface{}, client *apiClient) ([]map[string]interface{}, diag.Diagnostics) {
	var customSchemas []map[string]interface{}

	directoryService, diags := client.NewDirectoryService()
	if diags.HasError() {
		return nil, diags
	}

	schemaService, diags := GetSchemasService(directoryService)
	if diags.HasError() {
		return nil, diags
	}

	for schemaName, sv := range schemaAttrObj.(map[string]googleapi.RawMessage) {
		schemaDef, err := schemaService.Get(client.Customer, schemaName).Do()
		if err != nil {
			return nil, diag.FromErr(err)
		}

		schemaFieldMap := map[string]*directory.SchemaFieldSpec{}
		for _, schemaField := range schemaDef.Fields {
			schemaFieldMap[schemaField.FieldName] = schemaField
		}

		var schemaValuesObj map[string]interface{}

		err = json.Unmarshal(sv, &schemaValuesObj)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		schemaValues := map[string]interface{}{}
		for k, v := range schemaValuesObj {
			if _, ok := schemaFieldMap[k]; !ok {
				return nil, append(diags, diag.Diagnostic{
					Summary:  fmt.Sprintf("field name (%s) is not found in this schema definition (%s)", k, schemaName),
					Severity: diag.Warning,
				})
			}

			if schemaFieldMap[k].MultiValued {
				vals := []interface{}{}
				for _, item := range v.([]interface{}) {
					val, err := convertFieldValueType(schemaFieldMap[k].FieldType, item.(map[string]interface{})["value"])
					if err != nil {
						return nil, diag.FromErr(err)
					}
					vals = append(vals, val)
				}
				jsonVals, err := json.Marshal(vals)
				if err != nil {
					return nil, diag.FromErr(err)
				}
				schemaValues[k] = string(jsonVals)
			} else {
				val, err := convertFieldValueType(schemaFieldMap[k].FieldType, v)
				if err != nil {
					return nil, diag.FromErr(err)
				}

				jsonVal, err := json.Marshal(val)
				if err != nil {
					return nil, diag.FromErr(err)
				}
				schemaValues[k] = string(jsonVal)
			}
		}

		customSchemas = append(customSchemas, map[string]interface{}{
			"schema_name":   schemaName,
			"schema_values": schemaValues,
		})
	}

	return customSchemas, nil
}
