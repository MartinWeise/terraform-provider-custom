package provider

import (
	"context"
	"net/http"
	"os"
	"terraform-provider-custom/btp/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func New() provider.Provider {
	return NewWithClient(http.DefaultClient)
}

func NewWithClient(httpClient *http.Client) provider.Provider {
	return &clientProvider{httpClient: httpClient}
}

type clientProvider struct {
	httpClient          *http.Client
	betaFeaturesEnabled bool
}

func (p *clientProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `The custom Terraform provider for SAP BTP enables you to automate the provisioning, management and configuration of resources on [SAP Business Technology Platform](https://account.hana.ondemand.com/).

### Features

* Manage Credential Store Passwords
* Manage Business Partners
* Read SAML 2.0 Destination Trust Certificates

> \#pfeift.`,
		Attributes: map[string]schema.Attribute{
			"credstore_binding_parameters": schema.StringAttribute{
				MarkdownDescription: "The credential binding parameters of the Credential Store API. This can also be sourced from the `CREDSTORE_BINDING_PARAMETERS` environment variable.",
				Sensitive:           true,
				Optional:            true,
			},
			"destination_binding_parameters": schema.StringAttribute{
				MarkdownDescription: "The credential binding parameters of the Destination Service API. This can also be sourced from the `DESTINATION_BINDING_PARAMETERS` environment variable.",
				Sensitive:           true,
				Optional:            true,
			},
			"cslncore_binding_parameters": schema.StringAttribute{
				MarkdownDescription: "The credential binding parameters of the CSLN Core Application. This can also be sourced from the `CSLNCORE_BINDING_PARAMETERS` environment variable.",
				Sensitive:           true,
				Optional:            true,
			},
			"cslnparticipants_binding_parameters": schema.StringAttribute{
				MarkdownDescription: "The credential binding parameters of the CSLN Participants Application. This can also be sourced from the `CSLNPARTICIPANTS_BINDING_PARAMETERS` environment variable.",
				Sensitive:           true,
				Optional:            true,
			},
			"csln_domain": schema.StringAttribute{
				MarkdownDescription: "The domain of the CSLN applications. This can also be sourced from the `CSLN_DOMAIN` environment variable.",
				Sensitive:           true,
				Optional:            true,
			},
		},
	}
}

type providerData struct {
	CredStoreBindingParameters        types.String `tfsdk:"credstore_binding_parameters"`
	DestinationBindingParameters      types.String `tfsdk:"destination_binding_parameters"`
	CSLNCoreBindingParameters         types.String `tfsdk:"cslncore_binding_parameters"`
	CSLNDomain                        types.String `tfsdk:"csln_domain"`
	CSLNParticipantsBindingParameters types.String `tfsdk:"cslnparticipants_binding_parameters"`
}

// Metadata returns the provider type name.
func (p *clientProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "custom"
}

func (p *clientProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	const unableToCreateClient = "unableToCreateClient"

	// Retrieve provider data from configuration
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// User must provide Cloud Foundry Credential Store Binding Parameters to the provider
	var credStoreParameters string
	if config.CredStoreBindingParameters.IsUnknown() {
		resp.Diagnostics.AddWarning(unableToCreateClient, "Cannot use unknown value as credstore_binding_parameters")
		return
	}

	// favors environment variables over static config
	if !config.CredStoreBindingParameters.IsNull() {
		credStoreParameters = config.CredStoreBindingParameters.ValueString()
	} else {
		credStoreParameters = os.Getenv("CREDSTORE_BINDING_PARAMETERS")
	}

	var credStoreParams *client.CredStoreBindingParameters
	if len(credStoreParameters) != 0 {
		var err error
		credStoreParams, err = parseCredStoreParams(ctx, credStoreParameters)

		if err != nil {
			resp.Diagnostics.AddError(unableToCreateClient, "Cannot parse credential store parameters: "+err.Error())
			return
		}
	}

	// User must provide Destination Service Binding Parameters to the provider
	var destinationParameters string
	if config.DestinationBindingParameters.IsUnknown() {
		resp.Diagnostics.AddWarning(unableToCreateClient, "Cannot use unknown value as destination_binding_parameters")
		return
	}

	// favors environment variables over static config
	if !config.DestinationBindingParameters.IsNull() {
		destinationParameters = config.DestinationBindingParameters.ValueString()
	} else {
		destinationParameters = os.Getenv("DESTINATION_BINDING_PARAMETERS")
	}

	var destinationParams *client.OAuthTokenFlowBindingParameters
	if len(destinationParameters) != 0 {
		var err error
		destinationParams, err = parseDestinationParams(ctx, destinationParameters)

		if err != nil {
			resp.Diagnostics.AddError(unableToCreateClient, "Cannot parse destination parameters: "+err.Error())
			return
		}
	}

	// User must provide Destination Service Binding Parameters to the provider
	var cslnCoreParameters string
	if config.CSLNCoreBindingParameters.IsUnknown() {
		resp.Diagnostics.AddWarning(unableToCreateClient, "Cannot use unknown value as cslncore_binding_parameters")
		return
	}

	// favors environment variables over static config
	if !config.CSLNCoreBindingParameters.IsNull() {
		cslnCoreParameters = config.CSLNCoreBindingParameters.ValueString()
	} else {
		cslnCoreParameters = os.Getenv("CSLNCORE_BINDING_PARAMETERS")
	}

	// favors environment variables over static config
	var cslnDomain string
	if !config.CSLNDomain.IsNull() {
		cslnDomain = config.CSLNDomain.ValueString()
	} else {
		cslnDomain = os.Getenv("CSLN_DOMAIN")
	}

	var cslnCoreParams *client.OAuthTokenFlowBindingParameters
	if len(cslnCoreParameters) != 0 {
		var err error
		cslnCoreParams, err = parseCSLNCoreParams(ctx, cslnCoreParameters, cslnDomain)

		if err != nil {
			resp.Diagnostics.AddError(unableToCreateClient, "Cannot parse csln core parameters: "+err.Error())
			return
		}
	}

	// User must provide Destination Service Binding Parameters to the provider
	var cslnParticipantsParameters string
	if config.CSLNParticipantsBindingParameters.IsUnknown() {
		resp.Diagnostics.AddWarning(unableToCreateClient, "Cannot use unknown value as cslnparticipants_binding_parameters")
		return
	}

	// favors environment variables over static config
	if !config.CSLNParticipantsBindingParameters.IsNull() {
		cslnParticipantsParameters = config.CSLNParticipantsBindingParameters.ValueString()
	} else {
		cslnParticipantsParameters = os.Getenv("CSLNPARTICIPANTS_BINDING_PARAMETERS")
	}

	var cslnParticipantsParams *client.OAuthTokenFlowStubBindingParameters
	if len(cslnCoreParameters) != 0 {
		var err error
		cslnParticipantsParams, err = parseCSLNParticipantsParams(ctx, cslnParticipantsParameters, cslnDomain)

		if err != nil {
			resp.Diagnostics.AddError(unableToCreateClient, "Cannot parse csln participants parameters: "+err.Error())
			return
		}
	}

	btpClient := client.NewClientFacade(client.NewV2ClientWithHttpClient(p.httpClient, credStoreParams,
		destinationParams, cslnCoreParams, cslnParticipantsParams, req.TerraformVersion))

	if resp.Diagnostics.HasError() {
		return
	}

	resp.DataSourceData = btpClient
	resp.ResourceData = btpClient
}

// Resources - Defines provider resources
func (p *clientProvider) Resources(ctx context.Context) []func() resource.Resource {
	betaResources := []func() resource.Resource{
		//Beta resources should be excluded from sonar scan.
		//If you add them to production code, remove them from sonar exclusion list
	}

	if !p.betaFeaturesEnabled {
		betaResources = nil
	}

	return append([]func() resource.Resource{
		newBusinessPartnerCompanyGroup,
		newBusinessPartnerCompanyGroupContact,
		newCredentialStorePassword,
	}, betaResources...)
}

// DataSources - Defines provider data sources
func (p *clientProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	betaDataSources := []func() datasource.DataSource{
		//Beta data sources should be excluded from sonar scan.
		//If you add them to production code, remove them from sonar exclusion list

		//newDirectoryAppDataSource,
	}

	if !p.betaFeaturesEnabled {
		betaDataSources = nil
	}

	return append([]func() datasource.DataSource{
		newCredentialStorePasswordDataSource,
		newDestinationSamlMetadataDataSource,
		newBusinessPartnerCompanyGroupContactDataSource,
	}, betaDataSources...)
}

func (p *clientProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{
		NewExtractIdentityServiceUrlFunction,
		NewExtractIdpOriginFunction,
	}
}
