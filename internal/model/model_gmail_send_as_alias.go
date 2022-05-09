package model

import "github.com/hashicorp/terraform-plugin-framework/types"

type GmailSendAsAliasResourceData struct {
	ID                 types.String `tfsdk:"id"`
	PrimaryEmail       types.String `tfsdk:"primary_email"`
	SendAsEmail        types.String `tfsdk:"send_as_email"`
	DisplayName        types.String `tfsdk:"display_name"`
	ReplyToAddress     types.String `tfsdk:"reply_to_address"`
	Signature          types.String `tfsdk:"signature"`
	IsPrimary          types.Bool   `tfsdk:"is_primary"`
	IsDefault          types.Bool   `tfsdk:"is_default"`
	TreatAsAlias       types.Bool   `tfsdk:"treat_as_alias"`
	SmtpMsa            types.Object `tfsdk:"smtp_msa"`
	VerificationStatus types.String `tfsdk:"verification_status"`
}

type GmailSendAsAliasResourceSmtpMsa struct {
	Host         types.String `tfsdk:"host"`
	Port         types.Int64  `tfsdk:"port"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
	SecurityMode types.String `tfsdk:"security_mode"`
}
