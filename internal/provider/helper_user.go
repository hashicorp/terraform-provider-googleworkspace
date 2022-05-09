package googleworkspace

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-googleworkspace-pf/internal/model"
)

func emailsToInterfaces(ctx context.Context, planned []attr.Value, diags *diag.Diagnostics) []interface{} {
	var emails []interface{}
	for _, e := range planned {
		var em model.UserResourceEmail
		d := e.(types.Object).As(ctx, &em, types.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return []interface{}{}
		}

		emails = append(emails, map[string]interface{}{
			"address":    em.Address.Value,
			"customType": em.CustomType.Value,
			"primary":    em.Primary.Value,
			"type":       em.Type.Value,
		})
	}

	return emails
}

func externalIdsToInterfaces(ctx context.Context, planned []attr.Value, diags *diag.Diagnostics) []interface{} {
	var externalIds []interface{}
	for _, e := range planned {
		var ei model.UserResourceExternalId
		d := e.(types.Object).As(ctx, &ei, types.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return []interface{}{}
		}

		externalIds = append(externalIds, map[string]interface{}{
			"customType": ei.CustomType.Value,
			"type":       ei.Type.Value,
			"value":      ei.Value.Value,
		})
	}

	return externalIds
}

func relationsToInterfaces(ctx context.Context, planned []attr.Value, diags *diag.Diagnostics) []interface{} {
	var relations []interface{}
	for _, r := range planned {
		var rel model.UserResourceRelation
		d := r.(types.Object).As(ctx, &rel, types.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return []interface{}{}
		}

		relations = append(relations, map[string]interface{}{
			"customType": rel.CustomType.Value,
			"type":       rel.Type.Value,
			"value":      rel.Value.Value,
		})
	}

	return relations
}

func addressesToInterfaces(ctx context.Context, planned []attr.Value, diags *diag.Diagnostics) []interface{} {
	var addresses []interface{}
	for _, a := range planned {
		var addr model.UserResourceAddress
		d := a.(types.Object).As(ctx, &addr, types.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return []interface{}{}
		}

		addresses = append(addresses, map[string]interface{}{
			"country":            addr.Country.Value,
			"countryCode":        addr.CountryCode.Value,
			"customType":         addr.CustomType.Value,
			"extendedAddress":    addr.ExtendedAddress.Value,
			"formatted":          addr.Formatted.Value,
			"locality":           addr.Locality.Value,
			"poBox":              addr.PoBox.Value,
			"postalCode":         addr.PostalCode.Value,
			"primary":            addr.Primary.Value,
			"region":             addr.Region.Value,
			"sourceIsStructured": addr.SourceIsStructured.Value,
			"streetAddress":      addr.StreetAddress.Value,
			"type":               addr.Type.Value,
		})
	}

	return addresses
}

func organizationsToInterfaces(ctx context.Context, planned []attr.Value, diags *diag.Diagnostics) []interface{} {
	var organizations []interface{}
	for _, o := range planned {
		var org model.UserResourceOrganization
		d := o.(types.Object).As(ctx, &org, types.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return []interface{}{}
		}

		organizations = append(organizations, map[string]interface{}{
			"costCenter":         org.CostCenter.Value,
			"customType":         org.CustomType.Value,
			"department":         org.Department.Value,
			"description":        org.Description.Value,
			"domain":             org.Domain.Value,
			"fullTimeEquivalent": org.FullTimeEquivalent.Value,
			"location":           org.Location.Value,
			"name":               org.Name.Value,
			"primary":            org.Primary.Value,
			"symbol":             org.Symbol.Value,
			"title":              org.Title.Value,
			"type":               org.Type.Value,
		})
	}

	return organizations
}

func phonesToInterfaces(ctx context.Context, planned []attr.Value, diags *diag.Diagnostics) []interface{} {
	var phones []interface{}
	for _, p := range planned {
		var ph model.UserResourcePhone
		d := p.(types.Object).As(ctx, &ph, types.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return []interface{}{}
		}

		phones = append(phones, map[string]interface{}{
			"customType": ph.CustomType.Value,
			"primary":    ph.Primary.Value,
			"type":       ph.Type.Value,
			"value":      ph.Value.Value,
		})
	}

	return phones
}

func languagesToInterfaces(ctx context.Context, planned []attr.Value, diags *diag.Diagnostics) []interface{} {
	var languages []interface{}
	for _, l := range planned {
		var lang model.UserResourceLanguage
		d := l.(types.Object).As(ctx, &lang, types.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return []interface{}{}
		}

		languages = append(languages, map[string]interface{}{
			"customLanguage": lang.CustomLanguage.Value,
			"languageCode":   lang.LanguageCode.Value,
			"preference":     lang.Preference.Value,
		})
	}

	return languages
}

func posixAccountsToInterfaces(ctx context.Context, planned []attr.Value, diags *diag.Diagnostics) []interface{} {
	var posixAccounts []interface{}
	for _, p := range planned {
		var pa model.UserResourcePosixAccount
		d := p.(types.Object).As(ctx, &pa, types.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return []interface{}{}
		}

		posixAccounts = append(posixAccounts, map[string]interface{}{
			"accountId":           pa.AccountId.Value,
			"gecos":               pa.Gecos.Value,
			"gid":                 pa.Gid.Value,
			"homeDirectory":       pa.HomeDirectory.Value,
			"operatingSystemType": pa.OperatingSystemType.Value,
			"primary":             pa.Primary.Value,
			"shell":               pa.Shell.Value,
			"systemId":            pa.SystemId.Value,
			"uid":                 pa.Uid.Value,
			"username":            pa.Username.Value,
		})
	}

	return posixAccounts
}

func sshPublicKeysToInterfaces(ctx context.Context, planned []attr.Value, diags *diag.Diagnostics) []interface{} {
	var sshPublicKeys []interface{}
	for _, k := range planned {
		var key model.UserResourceSshPublicKey
		d := k.(types.Object).As(ctx, &key, types.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return []interface{}{}
		}

		sshPublicKeys = append(sshPublicKeys, map[string]interface{}{
			"expirationTimeUsec": key.ExpirationTimeUsec.Value,
			"fingerprint":        key.Fingerprint.Value,
			"key":                key.Key.Value,
		})
	}

	return sshPublicKeys
}

func websitesToInterfaces(ctx context.Context, planned []attr.Value, diags *diag.Diagnostics) []interface{} {
	var websites []interface{}
	for _, w := range planned {
		var web model.UserResourcePhone
		d := w.(types.Object).As(ctx, &web, types.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return []interface{}{}
		}

		websites = append(websites, map[string]interface{}{
			"customType": web.CustomType.Value,
			"primary":    web.Primary.Value,
			"type":       web.Type.Value,
			"value":      web.Value.Value,
		})
	}

	return websites
}

func locationsToInterfaces(ctx context.Context, planned []attr.Value, diags *diag.Diagnostics) []interface{} {
	var locations []interface{}
	for _, l := range planned {
		var loc model.UserResourceLocation
		d := l.(types.Object).As(ctx, &loc, types.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return []interface{}{}
		}

		locations = append(locations, map[string]interface{}{
			"area":         loc.Area.Value,
			"buildingId":   loc.BuildingId.Value,
			"customType":   loc.CustomType.Value,
			"deskCode":     loc.DeskCode.Value,
			"floorName":    loc.FloorName.Value,
			"floorSection": loc.FloorSection.Value,
			"type":         loc.Type.Value,
		})
	}

	return locations
}

func keywordsToInterfaces(ctx context.Context, planned []attr.Value, diags *diag.Diagnostics) []interface{} {
	var keywords []interface{}
	for _, k := range planned {
		var kw model.UserResourceRelation
		d := k.(types.Object).As(ctx, &kw, types.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return []interface{}{}
		}

		keywords = append(keywords, map[string]interface{}{
			"customType": kw.CustomType.Value,
			"type":       kw.Type.Value,
			"value":      kw.Value.Value,
		})
	}

	return keywords
}

func imsToInterfaces(ctx context.Context, planned []attr.Value, diags *diag.Diagnostics) []interface{} {
	var ims []interface{}
	for _, i := range planned {
		var im model.UserResourceIm
		d := i.(types.Object).As(ctx, &im, types.ObjectAsOptions{})
		diags.Append(d...)
		if diags.HasError() {
			return []interface{}{}
		}

		ims = append(ims, map[string]interface{}{
			"customProtocol": im.CustomProtocol.Value,
			"customType":     im.CustomType.Value,
			"im":             im.Im.Value,
			"primary":        im.Primary.Value,
			"protocol":       im.Protocol.Value,
			"type":           im.Type.Value,
		})
	}

	return ims
}

var customSchemas []interface{}
