package provider

import (
	"context"
	"fmt"
	"terraform-provider-custom/btp/client"
	"terraform-provider-custom/btp/client/btp"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func newDestinationSamlMetadataDataSource() datasource.DataSource {
	return &destinationSamlMetadataDataSource{}
}

type destinationSamlMetadataDataSource struct {
	cli *client.ClientFacade
}

func (rs *destinationSamlMetadataDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_destination_saml_metadata", req.ProviderTypeName)
}

func (rs *destinationSamlMetadataDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	rs.cli = req.ProviderData.(*client.ClientFacade)
}

func (rs *destinationSamlMetadataDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Get the SAML IdP metadata used to share configuration information between the Identity Provider (IdP) and the Service Provider (SP). Note that it's generic (contains only the certificate) and not configured for a specific scenario (i.e. CF to CF, CF to Neo etc.) which requires additional properties.

**Further documentation:**
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/set-up-trust-between-systems>`,
		Attributes: map[string]schema.Attribute{
			"idp_metadata": schema.StringAttribute{
				MarkdownDescription: "Base64-encoded SAML IdP metadata XML",
				Computed:            true,
			},
		},
	}
}

func (rs *destinationSamlMetadataDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data destinationTrustMetadataType
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := rs.cli.DestinationTrust.Metadata.Read(ctx)

	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("%s", err))
		return
	}

	raw, err := client.Parse[*btp.IdpMetadata](res.Body)

	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("%s", err))
		return
	}

	state, diags := destinationTrustMetadataFromValue(raw)
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
