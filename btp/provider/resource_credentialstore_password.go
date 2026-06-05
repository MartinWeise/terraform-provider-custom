package provider

import (
	"context"
	"fmt"
	"terraform-provider-custom/btp/client"
	"terraform-provider-custom/btp/client/btp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func newCredentialStorePassword() resource.Resource {
	return &credentialStorePassword{}
}

type credentialStorePassword struct {
	client *client.ClientFacade
}

func (rs *credentialStorePassword) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_credentialstore_password", req.ProviderTypeName)
}

func (rs *credentialStorePassword) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	rs.client = req.ProviderData.(*client.ClientFacade)
}

func (rs *credentialStorePassword) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Create or update password credential with the specified name in the specified namespace.

The authentication to the client API is implemented via basic credentials, and the payload encryption is mandatory enabled.

**Further documentation:**
<https://help.sap.com/docs/credential-store/sap-credential-store/sap-credential-store>`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier of the credential",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the credential.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "Namespace of the credential.",
				Required:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Value of the credential.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 4096),
				},
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "Optional attribute that can be used to store additional information about the credential. The value of the attribute is not processed by the service and is stored as is.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(10000),
				},
			},
			"unmodifiable": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether the attributes of the credential are able to be changed.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false), // FIXME BTP ignores <null> and defaults to <false>, resulting in an inconsistent state.
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username associated with the value.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(1024),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The credential type.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("password"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("password"),
				},
			},
			"modified_at": schema.StringAttribute{
				MarkdownDescription: "The timestamp when the credential was last modified.",
				Computed:            true,
			},
		},
	}
}

func (rs *credentialStorePassword) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state credentialStorePasswordType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := rs.client.CredentialStore.Password.ReadByNamespace(ctx, state.Namespace.ValueString(), state.Name.ValueString())

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

	raw, err := client.DecryptResponse(ctx, res.Body, *rs.client.CredStoreParams.EncryptionPrivateKey)

	if err != nil {
		resp.Diagnostics.AddError("Decrypt Error", fmt.Sprintf("%s", err))
		return
	}

	password, err := client.Parse[*btp.Password](raw)

	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("%s", err))
		return
	}

	updatedState, diags := credentialStorePasswordFromValue(password, state.Namespace.ValueString(), state.Value.ValueString())
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, &updatedState)
	resp.Diagnostics.Append(diags...)
}

func (rs *credentialStorePassword) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan credentialStorePasswordType
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := rs.client.CredentialStore.Password.CreateOrUpdateByNamespace(ctx, plan.Namespace.ValueString(), &btp.NewPassword{
		Name:         plan.Name.ValueString(),
		Value:        plan.Value.ValueString(),
		Metadata:     plan.Metadata.ValueStringPointer(),
		Unmodifiable: plan.Unmodifiable.ValueBool(),
		Username:     plan.Username.ValueStringPointer(),
	})

	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("%s", err))
		return
	}

	raw, err := client.DecryptResponse(ctx, res.Body, *rs.client.CredStoreParams.EncryptionPrivateKey)

	if err != nil {
		resp.Diagnostics.AddError("Decrypt Error", fmt.Sprintf("%s", err))
		return
	}

	password, err := client.Parse[*btp.Password](raw)

	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("%s", err))
		return
	}

	state, diags := credentialStorePasswordFromValue(password, plan.Namespace.ValueString(), plan.Value.ValueString())
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (rs *credentialStorePassword) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state credentialStorePasswordType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var plan credentialStorePasswordType
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := rs.client.CredentialStore.Password.CreateOrUpdateByNamespace(ctx, plan.Namespace.ValueString(), &btp.NewPassword{
		Name:         plan.Name.ValueString(),
		Value:        plan.Value.ValueString(),
		Metadata:     plan.Metadata.ValueStringPointer(),
		Unmodifiable: plan.Unmodifiable.ValueBool(),
		Username:     plan.Username.ValueStringPointer(),
	})

	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("%s", err))
		return
	}

	raw, err := client.DecryptResponse(ctx, res.Body, *rs.client.CredStoreParams.EncryptionPrivateKey)

	if err != nil {
		resp.Diagnostics.AddError("Decrypt Error", fmt.Sprintf("%s", err))
		return
	}

	password, err := client.Parse[*btp.Password](raw)

	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("%s", err))
		return
	}

	updatedState, diags := credentialStorePasswordFromValue(password, plan.Namespace.ValueString(), plan.Value.ValueString())
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, &updatedState)
	resp.Diagnostics.Append(diags...)
}

func (rs *credentialStorePassword) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state credentialStorePasswordType
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := rs.client.CredentialStore.Password.DeleteByNamespace(ctx, state.Namespace.ValueString(), state.Name.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("%s", err))
		return
	}
}
