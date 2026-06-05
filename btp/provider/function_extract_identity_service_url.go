package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

type ExtractIdentityServiceUrlFunction struct{}

var _ function.Function = &ExtractIdentityServiceUrlFunction{}

func NewExtractIdentityServiceUrlFunction() function.Function {
	return &ExtractIdentityServiceUrlFunction{}
}

func (f *ExtractIdentityServiceUrlFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "extract_identity_service_url"
}

func (f *ExtractIdentityServiceUrlFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:             "extract_identity_service_url",
		Description:         "Parses the Host URL of a Identity Service instance and returns the value.",
		MarkdownDescription: "Parses the Host URL of a Identity Service instance and returns the value.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "url",
				Description:         "Service Instance URL",
				MarkdownDescription: "Service Instance URL",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *ExtractIdentityServiceUrlFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var object string

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &object))

	if resp.Error != nil {
		return
	}

	value, err := ExtractHostname(ctx, object)

	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(err.Error()))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, value))
}
