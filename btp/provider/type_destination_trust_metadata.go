package provider

import (
	"terraform-provider-custom/btp/client/btp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type destinationTrustMetadataType struct {
	IdpMetadata types.String `tfsdk:"idp_metadata"`
}

func destinationTrustMetadataFromValue(obj *btp.IdpMetadata) (destinationTrustMetadataType, diag.Diagnostics) {
	var certificate destinationTrustMetadataType

	certificate.IdpMetadata = types.StringValue(obj.IdpMetadata)

	return certificate, nil
}
