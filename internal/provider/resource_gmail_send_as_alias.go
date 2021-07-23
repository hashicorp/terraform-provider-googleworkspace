package googleworkspace

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"google.golang.org/api/gmail/v1"
)

// TODO ensure a safe character is chosen to separate emails for a composite id,
// otherwise it will be difficult to support import
const sendAsIdSeparator = ":"

func resourceGmailSendAsAlias() *schema.Resource {
	return &schema.Resource{
		Description: "Role resource in the Terraform Googleworkspace provider.",

		CreateContext: resourceGmailSendAsAliasCreate,
		ReadContext:   resourceGmailSendAsAliasRead,
		UpdateContext: resourceGmailSendAsAliasUpdate,
		DeleteContext: resourceGmailSendAsAliasDelete,

		// TODO split email:sendAs id, is ':' a safe separator?
		// Importer: &schema.ResourceImporter{
		// 	StateContext: nil,
		// },

		Schema: map[string]*schema.Schema{
			"primary_email": {
				Description: "User's primary email address.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"send_as_email": {
				Description: "The email address that appears in the 'From:' header for mail sent using this alias.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"display_name": {
				Description: "A name that appears in the 'From:' header for mail sent using this alias. For custom 'from' addresses, when this is empty, Gmail will populate the 'From:' header with the name that is used for the primary address associated with the account. If the admin has disabled the ability for users to update their name format, requests to update this field for the primary login will silently fail.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"reply_to_address": {
				Description: "An optional email address that is included in a 'Reply-To:' header for mail sent using this alias. If this is empty, Gmail will not generate a 'Reply-To:' header.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"signature": {
				Description: "An optional HTML signature that is included in messages composed with this alias in the Gmail web UI. This signature is added to new emails only.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"is_primary": {
				Description: "Whether this address is the primary address used to login to the account. Every Gmail account has exactly one primary address, and it cannot be deleted from the collection of send-as aliases.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"is_default": {
				Description: "Whether this address is selected as the default 'From:' address in situations such as composing a new message or sending a vacation auto-reply. Every Gmail account has exactly one default send-as address, so the only legal value that clients may write to this field is true. Changing this from false to true for an address will result in this field becoming false for the other previous default address. To avoid drift with Terraform, please change the previous default's config to false after this change is applied and perform a refresh to synchronize with remote state.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"treat_as_alias": {
				Description: "Whether Gmail should treat this address as an alias for the user's primary email address. This setting only applies to custom 'from' aliases. See https://support.google.com/a/answer/1710338 for help on making this decision",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true, // mirrors the UI
			},
			"smtp_msa": {
				Description: "An optional SMTP service that will be used as an outbound relay for mail sent using this alias. If this is empty, outbound mail will be sent directly from Gmail's servers to the destination SMTP service. This setting only applies to custom 'from' aliases.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Description: "The hostname of the SMTP service.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"port": {
							Description: "The port of the SMTP service.",
							Type:        schema.TypeInt,
							Required:    true,
						},
						"username": {
							Description:  "The username that will be used for authentication with the SMTP service. This is a write-only field that can be specified in requests to create or update SendAs settings; it is never populated in responses.",
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"smtp_msa.0.password"},
						},
						"password": {
							Description: "The password that will be used for authentication with the SMTP service. This is a write-only field that can be specified in requests to create or update SendAs settings; it is never populated in responses.",
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
						},
						"security_mode": {
							Description:      "he protocol that will be used to secure communication with the SMTP service.",
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "securityModeUnspecified",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"securityModeUnspecified", "none", "ssl", "starttls"}, false)),
						},
					},
				},
			},
			"verification_status": {
				Description: "Indicates whether this address has been verified for use as a send-as alias.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceGmailSendAsAliasCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	gmailService, diags := client.NewGmailService()
	if diags.HasError() {
		return diags
	}

	sendAsAliasService, diags := GetGmailSendAsAliasService(gmailService)
	if diags.HasError() {
		return diags
	}

	primaryEmail := d.Get("primary_email").(string)
	sendAsEmail := d.Get("send_as_email").(string)
	log.Printf("[DEBUG] Creating Gmail Send As Alias %q", primaryEmail+sendAsIdSeparator+sendAsEmail)

	sendAs, err := sendAsAliasService.Create(primaryEmail, &gmail.SendAs{
		SendAsEmail:    sendAsEmail,
		DisplayName:    d.Get("display_name").(string),
		ReplyToAddress: d.Get("reply_to_address").(string),
		Signature:      d.Get("signature").(string),
		IsDefault:      d.Get("is_default").(bool),
		TreatAsAlias:   d.Get("treat_as_alias").(bool),
		SmtpMsa:        expandSmtpMsa(d.Get("smtp_msa").([]interface{})),
	}).Do()
	if err != nil {
		diag.FromErr(err)
	}

	d.Set("send_as_email", sendAs.SendAsEmail)
	d.SetId(primaryEmail + sendAsIdSeparator + sendAs.SendAsEmail)

	log.Printf("[DEBUG] Finished creating Gmail Send As Alias %q", d.Id())

	return resourceGmailSendAsAliasRead(ctx, d, meta)
}

func resourceGmailSendAsAliasUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	gmailService, diags := client.NewGmailService()
	if diags.HasError() {
		return diags
	}

	sendAsAliasService, diags := GetGmailSendAsAliasService(gmailService)
	if diags.HasError() {
		return diags
	}

	log.Printf("[DEBUG] Updating Gmail Send As Alias %q", d.Id())

	log.Printf("[DEBUG] Finished updating Gmail Send As Alias %q", d.Id())

	return resourceGmailSendAsAliasRead(ctx, d, meta)
}

func resourceGmailSendAsAliasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	gmailService, diags := client.NewGmailService()
	if diags.HasError() {
		return diags
	}

	sendAsAliasService, diags := GetGmailSendAsAliasService(gmailService)
	if diags.HasError() {
		return diags
	}

	log.Printf("[DEBUG] Getting Gmail Send As Alias %q", d.Id())

	primaryEmail := d.Get("primary_email").(string)
	sendAs, err := sendAsAliasService.Get(primaryEmail, d.Get("send_as_email").(string)).Do()
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished getting Gmail Send As Alias %q", d.Id())

	d.SetId(primaryEmail + sendAsIdSeparator + sendAs.SendAsEmail)
	d.Set("send_as_email", sendAs.SendAsEmail)
	d.Set("display_name", sendAs.DisplayName)
	d.Set("reply_to_address", sendAs.ReplyToAddress)
	d.Set("signature", sendAs.Signature)
	d.Set("is_primary", sendAs.IsPrimary)
	d.Set("is_default", sendAs.IsDefault)
	d.Set("treat_as_alias", sendAs.TreatAsAlias)
	d.Set("verification_status", sendAs.VerificationStatus)
	if sendAs.SmtpMsa != nil {
		if err := d.Set("smtp_msa", flattenSmtpMsa(sendAs.SmtpMsa, d)); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceGmailSendAsAliasDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	gmailService, diags := client.NewGmailService()
	if diags.HasError() {
		return diags
	}

	sendAsAliasService, diags := GetGmailSendAsAliasService(gmailService)
	if diags.HasError() {
		return diags
	}

	log.Printf("[DEBUG] Deleting Gmail Send As Alias %q", d.Id())

	log.Printf("[DEBUG] Finished deleting Gmail Send As Alias %q", d.Id())

	return nil
}

func expandSmtpMsa(smtpMsa []interface{}) *gmail.SmtpMsa {
	if len(smtpMsa) == 0 {
		return nil
	}
	values := smtpMsa[0].(map[string]interface{})
	return &gmail.SmtpMsa{
		Host:         values["host"].(string),
		Port:         int64(values["port"].(int)),
		Username:     values["username"].(string),
		Password:     values["password"].(string),
		SecurityMode: values["security_mode"].(string),
	}
}

func flattenSmtpMsa(smtpMsa *gmail.SmtpMsa, d *schema.ResourceData) []interface{} {
	result := make(map[string]interface{})

	// need to retrieve username/password from config
	configSmtpMsa := expandSmtpMsa(d.Get("smtp_msa").([]interface{}))

	result["host"] = smtpMsa.Host
	result["port"] = int(smtpMsa.Port)
	result["security_mode"] = smtpMsa.SecurityMode
	result["username"] = configSmtpMsa.Username
	result["password"] = configSmtpMsa.Password

	return []interface{}{result}
}
