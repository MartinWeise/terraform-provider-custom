package provider

import (
	"terraform-provider-custom/btp/client/sap"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type businessPartnerCompanyGroupType struct {
	Id        types.String `tfsdk:"id"`
	DisplayId types.String `tfsdk:"display_id"`
	TenantId  types.String `tfsdk:"tenant_id"`
	Subdomain types.String `tfsdk:"subdomain"`
}

func businessPartnerCompanyGroupFromValue(obj *sap.BusinessPartner) (businessPartnerCompanyGroupType, diag.Diagnostics) {
	var companyGroup businessPartnerCompanyGroupType

	companyGroup.Id = types.StringValue(obj.Id)
	companyGroup.DisplayId = types.StringValue(obj.DisplayId)

	return companyGroup, nil
}
