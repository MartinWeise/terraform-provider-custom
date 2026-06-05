package provider

import (
	"context"
	"fmt"
	"terraform-provider-custom/btp/client"
	"terraform-provider-custom/btp/client/btp"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func newBusinessPartnerCompanyGroupContactDataSource() datasource.DataSource {
	return &businessPartnerCompanyGroupContactDataSource{}
}

type businessPartnerCompanyGroupContactDataSource struct {
	rest *client.ClientFacade
}

func (rs *businessPartnerCompanyGroupContactDataSource) Metadata(_ context.Context, req datasource.MetadataRequest,
	resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_businesspartner_companygroup_contact", req.ProviderTypeName)
}

func (rs *businessPartnerCompanyGroupContactDataSource) Configure(_ context.Context, req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	rs.rest = req.ProviderData.(*client.ClientFacade)
}

func (rs *businessPartnerCompanyGroupContactDataSource) Schema(_ context.Context, _ datasource.SchemaRequest,
	resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Read a business partner company group contact using the SAP personal data manager.`,
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

func (rs *businessPartnerCompanyGroupContactDataSource) Read(ctx context.Context, req datasource.ReadRequest,
	resp *datasource.ReadResponse) {
	var data businessPartnerContactType
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	res, err := rs.rest.BusinessPartner.CompanyGroupContact.ReadByDisplayId(ctx, data.DisplayId.ValueString(), data.TenantSubdomain.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("%s", err))
		return
	}

	raw, err := client.Parse[*btp.BusinessPartnerContactResponse](res.Body)

	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("%s", err))
		return
	}

	if len(raw.Value) != 1 {
		resp.Diagnostics.AddError("CSLN Tenant Application Error", fmt.Sprintf("Company Group Contact with displayId %s not found in the CSLN Core application", data.DisplayId.ValueString()))
		return
	}

	val := raw.Value[0]

	state, diags := businessPartnerContactFromPValue(&val)
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
