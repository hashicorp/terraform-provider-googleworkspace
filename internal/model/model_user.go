package model

import "github.com/hashicorp/terraform-plugin-framework/types"

type UserResourceData struct {
	ID                            types.String `tfsdk:"id"`
	PrimaryEmail                  types.String `tfsdk:"primary_email"`
	Password                      types.String `tfsdk:"password"`
	HashFunction                  types.String `tfsdk:"hash_function"`
	IsAdmin                       types.Bool   `tfsdk:"is_admin"`
	IsDelegatedAdmin              types.Bool   `tfsdk:"is_delegated_admin"`
	AgreedToTerms                 types.Bool   `tfsdk:"agreed_to_terms"`
	Suspended                     types.Bool   `tfsdk:"suspended"`
	ChangePasswordAtNextLogin     types.Bool   `tfsdk:"change_password_at_next_login"`
	IpAllowlist                   types.Bool   `tfsdk:"ip_allowlist"`
	Name                          types.Object `tfsdk:"name"`
	Emails                        types.List   `tfsdk:"emails"`
	ExternalIds                   types.List   `tfsdk:"external_ids"`
	Relations                     types.List   `tfsdk:"relations"`
	Aliases                       types.List   `tfsdk:"aliases"`
	IsMailboxSetup                types.Bool   `tfsdk:"is_mailbox_setup"`
	CustomerId                    types.String `tfsdk:"customer_id"`
	Addresses                     types.List   `tfsdk:"addresses"`
	Organizations                 types.List   `tfsdk:"organizations"`
	LastLoginTime                 types.String `tfsdk:"last_login_time"`
	Phones                        types.List   `tfsdk:"phones"`
	SuspensionReason              types.String `tfsdk:"suspension_reason"`
	ThumbnailPhotoUrl             types.String `tfsdk:"thumbnail_photo_url"`
	Languages                     types.List   `tfsdk:"languages"`
	PosixAccounts                 types.List   `tfsdk:"posix_accounts"`
	CreationTime                  types.String `tfsdk:"creation_time"`
	NonEditableAliases            types.List   `tfsdk:"non_editable_aliases"`
	SshPublicKeys                 types.List   `tfsdk:"ssh_public_keys"`
	Websites                      types.List   `tfsdk:"websites"`
	Locations                     types.List   `tfsdk:"locations"`
	IncludeInGlobalAddressList    types.Bool   `tfsdk:"include_in_global_address_list"`
	Keywords                      types.List   `tfsdk:"keywords"`
	DeletionTime                  types.String `tfsdk:"deletion_time"`
	Ims                           types.List   `tfsdk:"ims"`
	CustomSchemas                 types.List   `tfsdk:"custom_schemas"`
	IsEnrolledIn2StepVerification types.Bool   `tfsdk:"is_enrolled_in_2_step_verification"`
	IsEnforcedIn2StepVerification types.Bool   `tfsdk:"is_enforced_in_2_step_verification"`
	Archived                      types.Bool   `tfsdk:"archived"`
	OrgUnitPath                   types.String `tfsdk:"org_unit_path"`
	RecoveryEmail                 types.String `tfsdk:"recovery_email"`
	RecoveryPhone                 types.String `tfsdk:"recovery_phone"`
}

type UserResourceName struct {
	FullName   types.String `tfsdk:"full_name"`
	FamilyName types.String `tfsdk:"family_name"`
	GivenName  types.String `tfsdk:"given_name"`
}

type UserResourceEmail struct {
	Address    types.String `tfsdk:"address"`
	CustomType types.String `tfsdk:"custom_type"`
	Primary    types.Bool   `tfsdk:"primary"`
	Type       types.String `tfsdk:"type"`
}

type UserResourceExternalId struct {
	CustomType types.String `tfsdk:"custom_type"`
	Type       types.String `tfsdk:"type"`
	Value      types.String `tfsdk:"value"`
}

type UserResourceRelation struct {
	CustomType types.String `tfsdk:"custom_type"`
	Type       types.String `tfsdk:"type"`
	Value      types.String `tfsdk:"value"`
}

type UserResourceAddress struct {
	Country            types.String `tfsdk:"country"`
	CountryCode        types.String `tfsdk:"country_code"`
	CustomType         types.String `tfsdk:"custom_type"`
	ExtendedAddress    types.String `tfsdk:"extended_address"`
	Formatted          types.String `tfsdk:"formatted"`
	Locality           types.String `tfsdk:"locality"`
	PoBox              types.String `tfsdk:"po_box"`
	PostalCode         types.String `tfsdk:"postal_code"`
	Primary            types.Bool   `tfsdk:"primary"`
	Region             types.String `tfsdk:"region"`
	SourceIsStructured types.Bool   `tfsdk:"source_is_structured"`
	StreetAddress      types.String `tfsdk:"street_address"`
	Type               types.String `tfsdk:"type"`
}

type UserResourceOrganization struct {
	CostCenter         types.String `tfsdk:"cost_center"`
	CustomType         types.String `tfsdk:"custom_type"`
	Department         types.String `tfsdk:"department"`
	Description        types.String `tfsdk:"description"`
	Domain             types.String `tfsdk:"domain"`
	FullTimeEquivalent types.String `tfsdk:"full_time_equivalent"`
	Location           types.String `tfsdk:"location"`
	Name               types.String `tfsdk:"name"`
	Primary            types.Bool   `tfsdk:"primary"`
	Symbol             types.String `tfsdk:"symbol"`
	Title              types.String `tfsdk:"title"`
	Type               types.String `tfsdk:"type"`
}

type UserResourcePhone struct {
	CustomType types.String `tfsdk:"custom_type"`
	Primary    types.Bool   `tfsdk:"primary"`
	Type       types.String `tfsdk:"type"`
	Value      types.String `tfsdk:"value"`
}

type UserResourceLanguage struct {
	CustomLanguage types.String `tfsdk:"custom_language"`
	LanguageCode   types.String `tfsdk:"language_code"`
	Preference     types.String `tfsdk:"preference"`
}

type UserResourcePosixAccount struct {
	AccountId           types.String `tfsdk:"account_id"`
	Gecos               types.String `tfsdk:"gecos"`
	Gid                 types.String `tfsdk:"gid"`
	HomeDirectory       types.String `tfsdk:"home_directory"`
	OperatingSystemType types.String `tfsdk:"operating_system_type"`
	Primary             types.Bool   `tfsdk:"primary"`
	Shell               types.String `tfsdk:"shell"`
	SystemId            types.String `tfsdk:"system_id"`
	Uid                 types.String `tfsdk:"uid"`
	Username            types.String `tfsdk:"username"`
}

type UserResourceSshPublicKey struct {
	ExpirationTimeUsec types.String `tfsdk:"expiration_time_usec"`
	Fingerprint        types.String `tfsdk:"fingerprint"`
	Key                types.String `tfsdk:"key"`
}

type UserResourceWebsite struct {
	CustomType types.String `tfsdk:"custom_type"`
	Primary    types.Bool   `tfsdk:"primary"`
	Type       types.String `tfsdk:"type"`
	Value      types.String `tfsdk:"value"`
}

type UserResourceLocation struct {
	Area         types.String `tfsdk:"area"`
	BuildingId   types.String `tfsdk:"building_id"`
	CustomType   types.String `tfsdk:"custom_type"`
	DeskCode     types.String `tfsdk:"desk_code"`
	FloorName    types.String `tfsdk:"floor_name"`
	FloorSection types.String `tfsdk:"floor_section"`
	Type         types.String `tfsdk:"type"`
}

type UserResourceKeyword struct {
	CustomType types.String `tfsdk:"custom_type"`
	Type       types.String `tfsdk:"type"`
	Value      types.String `tfsdk:"value"`
}

type UserResourceIm struct {
	CustomProtocol types.String `tfsdk:"custom_protocol"`
	CustomType     types.String `tfsdk:"custom_type"`
	Im             types.String `tfsdk:"im"`
	Primary        types.Bool   `tfsdk:"primary"`
	Protocol       types.String `tfsdk:"protocol"`
	Type           types.String `tfsdk:"type"`
}

type UserResourceCustomSchema struct {
	SchemaName   types.String `tfsdk:"schema_name"`
	SchemaValues types.Map    `tfsdk:"schema_values"`
}
