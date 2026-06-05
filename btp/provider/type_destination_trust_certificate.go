package provider

import (
	"terraform-provider-custom/btp/client/btp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type destinationTrustCertificateType struct {
	Certificate types.String `tfsdk:"certificate"`
}

func destinationTrustCertificateFromValue(obj *btp.TrustCertificateResponse) (destinationTrustCertificateType, diag.Diagnostics) {
	var certificate destinationTrustCertificateType

	certificate.Certificate = types.StringValue(obj.Certificate)

	return certificate, nil
}
