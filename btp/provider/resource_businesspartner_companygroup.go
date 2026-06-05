package provider

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-custom/btp/client"
	"terraform-provider-custom/btp/client/btp"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func newBusinessPartnerCompanyGroup() resource.Resource {
	return &businessPartnerCompanyGroup{}
}

type businessPartnerCompanyGroup struct {
	client *client.ClientFacade
}

func (rs *businessPartnerCompanyGroup) Metadata(_ context.Context, req resource.MetadataRequest,
	resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_business_partner_company_group", req.ProviderTypeName)
}

func (rs *businessPartnerCompanyGroup) Configure(_ context.Context, req resource.ConfigureRequest,
	_ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	rs.client = req.ProviderData.(*client.ClientFacade)
}

func (rs *businessPartnerCompanyGroup) Schema(_ context.Context, _ resource.SchemaRequest,
	resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Create or update the Company Group in the Business Partner application.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Globally-unique identifier from the database",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_id": schema.StringAttribute{
				MarkdownDescription: "Human-readable identifier for the Company Group",
				Required:            true,
			},
			"subdomain": schema.StringAttribute{
				MarkdownDescription: "Subdomain",
				Required:            true,
			},
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "Globally-unique identifier of the BTP subaccount",
				Required:            true,
			},
		},
	}
}

type BusinessPartnerCompanyGroupIdentityModel struct {
	DisplayId types.String `tfsdk:"display_id"`
	Subdomain types.String `tfsdk:"subdomain"`
	TenantId  types.String `tfsdk:"tenant_id"`
}

func (rs *businessPartnerCompanyGroup) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest,
	resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"display_id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"subdomain": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"tenant_id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

func (rs *businessPartnerCompanyGroup) Read(ctx context.Context, req resource.ReadRequest,
	resp *resource.ReadResponse) {
	var state businessPartnerCompanyGroupType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := rs.client.BusinessPartner.CompanyGroup.ReadByDisplayId(ctx, state.DisplayId.ValueString())

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

	raw, err := client.Parse[*btp.BusinessPartnerResponse](res.Body)

	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("%s", err))
		return
	}

	if len(raw.Value) != 1 {
		resp.Diagnostics.AddError("CSLN Core Application Error", fmt.Sprintf("Company Group with displayId %s not found in the CSLN Core application", state.DisplayId))
		return
	}

	val := raw.Value[0]

	updatedState, diags := businessPartnerCompanyGroupFromValue(&val)
	resp.Diagnostics.Append(diags...)

	updatedState.Subdomain = state.Subdomain
	updatedState.TenantId = state.TenantId

	diags = resp.State.Set(ctx, &updatedState)
	resp.Diagnostics.Append(diags...)

	var identity BusinessPartnerCompanyGroupIdentityModel

	diags = req.Identity.Get(ctx, &identity)
	if diags.HasError() {
		identity = BusinessPartnerCompanyGroupIdentityModel{
			DisplayId: types.StringValue(state.DisplayId.ValueString()),
			Subdomain: types.StringValue(state.Subdomain.ValueString()),
			TenantId:  types.StringValue(state.TenantId.ValueString()),
		}

		diags = resp.Identity.Set(ctx, identity)
		resp.Diagnostics.Append(diags...)
	}
}

func (rs *businessPartnerCompanyGroup) Create(ctx context.Context, req resource.CreateRequest,
	resp *resource.CreateResponse) {
	var plan businessPartnerCompanyGroupType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get Business Partner UUID (1)

	res, err := rs.client.BusinessPartner.CompanyGroup.ReadByDisplayId(ctx, plan.DisplayId.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("%s", err))
		return
	}

	raw, err := client.Parse[*btp.BusinessPartnerResponse](res.Body)

	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("%s", err))
		return
	}

	if len(raw.Value) != 1 {
		resp.Diagnostics.AddError("CSLN Core Application Error", fmt.Sprintf("Company Group with displayId %s not found in the CSLN Core application", plan.DisplayId.ValueString()))
		return
	}

	val := raw.Value[0]

	tflog.Debug(ctx, "Got business partner uuid", map[string]any{
		"id": val.Id,
	})

	// Update Subdomain (2)

	res, err = rs.client.BusinessPartner.CompanyGroup.UpdateSubdomain(ctx, plan.DisplayId.ValueString(), plan.Subdomain.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("%s", err))
		return
	}

	if res.StatusCode != 204 {
		resp.Diagnostics.AddError("CSLN Core Application Error", fmt.Sprintf("Could not update subdomain %s in the CSLN Core application", plan.Subdomain.ValueString()))
		return
	}

	tflog.Debug(ctx, "Updated subdomain", map[string]any{
		"id":        plan.Id.ValueString(),
		"subdomain": plan.Subdomain.ValueString(),
	})

	// Set Tenant UUID (3)

	res, err = rs.client.BusinessPartner.CompanyGroup.UpdateTenantUuid(ctx, plan.DisplayId.ValueString(), plan.TenantId.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("%s", err))
		return
	}

	if res.StatusCode != 204 {
		resp.Diagnostics.AddError("CSLN Core Application Error", fmt.Sprintf("Could not update tenant uuid %s in the CSLN Core application", plan.TenantId.ValueString()))
		return
	}

	tflog.Debug(ctx, "Updated tenant uuid", map[string]any{
		"id":          plan.Id.ValueString(),
		"tenant_uuid": plan.TenantId.ValueString(),
	})

	updatedState, diags := businessPartnerCompanyGroupFromValue(&val)
	resp.Diagnostics.Append(diags...)

	updatedState.Subdomain = plan.Subdomain
	updatedState.TenantId = plan.TenantId

	diags = resp.State.Set(ctx, &updatedState)
	resp.Diagnostics.Append(diags...)

	identity := BusinessPartnerCompanyGroupIdentityModel{
		DisplayId: types.StringValue(plan.DisplayId.ValueString()),
		Subdomain: types.StringValue(plan.Subdomain.ValueString()),
		TenantId:  types.StringValue(plan.TenantId.ValueString()),
	}

	diags = resp.Identity.Set(ctx, identity)
	resp.Diagnostics.Append(diags...)
}

func (rs *businessPartnerCompanyGroup) Delete(ctx context.Context, req resource.DeleteRequest,
	resp *resource.DeleteResponse) {
	panic("implement me")
}

func (rs *businessPartnerCompanyGroup) Update(ctx context.Context, request resource.UpdateRequest,
	response *resource.UpdateResponse) {
	panic("implement me")
}

// Create the function for the state import
func (rs *businessPartnerCompanyGroup) ImportState(ctx context.Context, req resource.ImportStateRequest,
	resp *resource.ImportStateResponse) {
	if req.ID != "" {
		idParts := strings.Split(req.ID, ",")

		if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
			resp.Diagnostics.AddError(
				"Unexpected Import Identifier",
				fmt.Sprintf("Expected import identifier with format: display_id, subdomain, tenant_id. Got: %q", req.ID),
			)
			return
		}

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("display_id"), idParts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subdomain"), idParts[1])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), idParts[2])...)
		return
	}

	var identityData BusinessPartnerCompanyGroupIdentityModel
	resp.Diagnostics.Append(req.Identity.Get(ctx, &identityData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("display_id"), identityData.DisplayId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("subdomain"), identityData.Subdomain)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), identityData.TenantId)...)
}
