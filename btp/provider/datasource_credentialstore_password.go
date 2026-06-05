package provider

import (
	"context"
	"fmt"
	"terraform-provider-custom/btp/client"
	"terraform-provider-custom/btp/client/btp"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func newCredentialStorePasswordDataSource() datasource.DataSource {
	return &credentialStorePasswordDataSource{}
}

type credentialStorePasswordDataSource struct {
	rest *client.ClientFacade
}

func (rs *credentialStorePasswordDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_credentialstore_password", req.ProviderTypeName)
}

func (rs *credentialStorePasswordDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	rs.rest = req.ProviderData.(*client.ClientFacade)
}

func (rs *credentialStorePasswordDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Read a password credential with the specified name in the specified namespace.

The authentication to the REST API is implemented via basic credentials, and the payload encryption is mandatory enabled.

**Further documentation:**
<https://help.sap.com/docs/credential-store/sap-credential-store/sap-credential-store>`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the credential",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the credential.",
				Computed:            true,
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "Namespace of the credential.",
				Computed:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Value of the credential.",
				Computed:            true,
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "Optional attribute that can be used to store additional information about the credential. The value of the attribute is not processed by the service and is stored as is.",
				Computed:            true,
			},
			"unmodifiable": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether the attributes of the credential are able to be changed.",
				Computed:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username associated with the value.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The credential type.",
				Computed:            true,
			},
			"modified_at": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the credential was last modified.",
				Computed:            true,
			},
		},
	}
}

func (rs *credentialStorePasswordDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data credentialStorePasswordType
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := rs.rest.CredentialStore.Password.ReadByNamespace(ctx, data.Namespace.ValueString(), data.Name.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("%s", err))
		return
	}

	raw, err := client.DecryptResponse(ctx, res.Body, *rs.rest.CredStoreParams.EncryptionPrivateKey)

	if err != nil {
		resp.Diagnostics.AddError("Decrypt Error", fmt.Sprintf("%s", err))
		return
	}

	password, err := client.Parse[*btp.Password](raw)

	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("%s", err))
		return
	}

	state, diags := credentialStorePasswordFromValue(password, data.Namespace.ValueString(), data.Value.ValueString())
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
