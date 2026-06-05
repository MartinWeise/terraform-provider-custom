package provider

import (
	"terraform-provider-custom/btp/client/btp"
	"terraform-provider-custom/btp/client/sap"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type businessPartnerContactType struct {
	TenantSubdomain types.String `tfsdk:"tenant_subdomain"`
	DisplayId       types.String `tfsdk:"display_id"`
	GivenName       types.String `tfsdk:"given_name"`
	FamilyName      types.String `tfsdk:"family_name"`
	EmailAddress    types.String `tfsdk:"email_address"`
	Phone           types.String `tfsdk:"phone"`
	Mobile          types.String `tfsdk:"mobile"`
}

func businessPartnerContactFromValue(obj *btp.OnboardContact) (businessPartnerContactType, diag.Diagnostics) {
	var contact businessPartnerContactType

	contact.DisplayId = types.StringValue(obj.DisplayId)
	contact.GivenName = types.StringValue(obj.GivenName)
	contact.FamilyName = types.StringValue(obj.FamilyName)
	contact.EmailAddress = types.StringValue(obj.Email)
	contact.Phone = types.StringValue(obj.Phone)
	contact.Mobile = types.StringValue(obj.Mobile)

	return contact, nil
}

func businessPartnerContactFromPValue(obj *sap.CompanyGroupContact) (businessPartnerContactType, diag.Diagnostics) {
	var contact businessPartnerContactType

	contact.DisplayId = types.StringValue(obj.DisplayId)
	contact.GivenName = types.StringValue(obj.FirstName)
	contact.FamilyName = types.StringValue(obj.LastName)
	contact.EmailAddress = types.StringValue(obj.Email)
	contact.Phone = types.StringValue(obj.Phone)
	contact.Mobile = types.StringValue(obj.Mobile)

	return contact, nil
}
