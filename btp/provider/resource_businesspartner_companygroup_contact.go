package provider

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-custom/btp/client"
	"terraform-provider-custom/btp/client/btp"
	"terraform-provider-custom/btp/client/sap"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newBusinessPartnerCompanyGroupContact() resource.Resource {
	return &businessPartnerCompanyGroupContact{}
}

type businessPartnerCompanyGroupContact struct {
	client *client.ClientFacade
}

func (rs *businessPartnerCompanyGroupContact) Metadata(_ context.Context, req resource.MetadataRequest,
	resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_business_partner_company_group_contact", req.ProviderTypeName)
}

func (rs *businessPartnerCompanyGroupContact) Configure(_ context.Context, req resource.ConfigureRequest,
	_ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	rs.client = req.ProviderData.(*client.ClientFacade)
}

func (rs *businessPartnerCompanyGroupContact) Schema(_ context.Context, _ resource.SchemaRequest,
	resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Create or update the Company Group in the Business Partner application.`,
		Attributes: map[string]schema.Attribute{
			"display_id": schema.StringAttribute{
				MarkdownDescription: "Human-readable unique identifier",
				Required:            true,
			},
			"tenant_subdomain": schema.StringAttribute{
				MarkdownDescription: "Tenant subdomain",
				Required:            true,
			},
			"given_name": schema.StringAttribute{
				MarkdownDescription: "Firstname",
				Required:            true,
			},
			"family_name": schema.StringAttribute{
				MarkdownDescription: "Lastname",
				Required:            true,
			},
			"email_address": schema.StringAttribute{
				MarkdownDescription: "Unique e-mail address that receives the onboarding email to set a password",
				Required:            true,
			},
			"phone": schema.StringAttribute{
				MarkdownDescription: "Phone number",
				Required:            true,
			},
			"mobile": schema.StringAttribute{
				MarkdownDescription: "Mobile phone number",
				Required:            true,
			},
		},
	}
}

type BusinessPartnerCompanyGroupContactIdentityModel struct {
	DisplayId       types.String `tfsdk:"display_id"`
	TenantSubdomain types.String `tfsdk:"tenant_subdomain"`
}

func (rs *businessPartnerCompanyGroupContact) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest,
	resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"display_id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"tenant_subdomain": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

func (rs *businessPartnerCompanyGroupContact) Read(ctx context.Context, req resource.ReadRequest,
	resp *resource.ReadResponse) {
	var state businessPartnerContactType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := rs.client.BusinessPartner.CompanyGroupContact.ReadByDisplayId(ctx, state.DisplayId.ValueString(), state.TenantSubdomain.ValueString())

	if err != nil {
		// Treat HTTP 404 Not Found status as a signal to recreate resource
		// #ref https://developer.hashicorp.com/terraform/plugin/framework/resources/read#recommendations
		if res != nil && res.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("%s", err))
		return
	}

	raw, err := client.Parse[*btp.BusinessPartnerContactResponse](res.Body)

	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("%s", err))
		return
	}

	if len(raw.Value) != 1 {
		resp.Diagnostics.AddError("CSLN Tenant Application Error", fmt.Sprintf("Company Group Contact with displayId %s not found in the CSLN Core application", state.DisplayId))
		return
	}

	val := raw.Value[0]

	updatedState, diags := businessPartnerContactFromPValue(&val)

	updatedState.TenantSubdomain = state.TenantSubdomain

	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	var identity BusinessPartnerCompanyGroupContactIdentityModel

	diags = req.Identity.Get(ctx, &identity)
	if diags.HasError() {
		identity = BusinessPartnerCompanyGroupContactIdentityModel{
			DisplayId:       types.StringValue(state.DisplayId.ValueString()),
			TenantSubdomain: types.StringValue(state.TenantSubdomain.ValueString()),
		}

		diags = resp.Identity.Set(ctx, identity)
		resp.Diagnostics.Append(diags...)
	}
}

func (rs *businessPartnerCompanyGroupContact) Create(ctx context.Context, req resource.CreateRequest,
	resp *resource.CreateResponse) {
	var plan businessPartnerContactType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Onboard Contact (4)

	res, err := rs.client.BusinessPartner.CompanyGroupContact.OnboardContact(ctx, plan.TenantSubdomain.ValueString(), &sap.OnboardContact{
		DisplayId:  plan.DisplayId.ValueString(),
		GivenName:  plan.GivenName.ValueString(),
		FamilyName: plan.FamilyName.ValueString(),
		Email:      plan.EmailAddress.ValueString(),
		Phone:      plan.Phone.ValueString(),
		Mobile:     plan.Mobile.ValueString(),
	})

	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("%s", err))
		return
	}

	raw, err := client.Parse[*btp.ValueResponse](res.Body)

	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("%s", err))
		return
	}

	updatedState, diags := businessPartnerContactFromValue(&raw.Value)

	updatedState.TenantSubdomain = plan.TenantSubdomain

	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, &updatedState)
	resp.Diagnostics.Append(diags...)

	identity := BusinessPartnerCompanyGroupContactIdentityModel{
		DisplayId:       types.StringValue(plan.DisplayId.ValueString()),
		TenantSubdomain: types.StringValue(plan.TenantSubdomain.ValueString()),
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
}

func (rs *businessPartnerCompanyGroupContact) Delete(ctx context.Context, req resource.DeleteRequest,
	resp *resource.DeleteResponse) {
	panic("implement me")
}

func (rs *businessPartnerCompanyGroupContact) Update(ctx context.Context, request resource.UpdateRequest,
	response *resource.UpdateResponse) {
	panic("implement me")
}

// Create the function for the state import
func (rs *businessPartnerCompanyGroupContact) ImportState(ctx context.Context, req resource.ImportStateRequest,
	resp *resource.ImportStateResponse) {
	if req.ID != "" {
		idParts := strings.Split(req.ID, ",")

		if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
			resp.Diagnostics.AddError(
				"Unexpected Import Identifier",
				fmt.Sprintf("Expected import identifier with format: display_id, tenant_subdomain. Got: %q", req.ID),
			)
			return
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("display_id"), idParts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_subdomain"), idParts[1])...)
		return
	}

	var identityData BusinessPartnerCompanyGroupContactIdentityModel
	resp.Diagnostics.Append(req.Identity.Get(ctx, &identityData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("display_id"), identityData.DisplayId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_subdomain"), identityData.TenantSubdomain)...)
}
