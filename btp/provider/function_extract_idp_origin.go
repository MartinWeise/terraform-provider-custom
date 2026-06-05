package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

type ExtractIdpOriginFunction struct{}

var _ function.Function = &ExtractIdpOriginFunction{}

func NewExtractIdpOriginFunction() function.Function {
	return &ExtractIdpOriginFunction{}
}

func (f *ExtractIdpOriginFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "extract_idp_origin"
}

func (f *ExtractIdpOriginFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:             "extract_idp_origin",
		Description:         "Parses the Origin a SAML 2.0 Identity Provider and returns the value.",
		MarkdownDescription: "Parses the Origin a SAML 2.0 Identity Provider and returns the value.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "idp_metadata",
				Description:         "Identity Provider Metadata as base64-encoded string",
				MarkdownDescription: "Identity Provider Metadata as base64-encoded string",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *ExtractIdpOriginFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var object string

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &object))

	if resp.Error != nil {
		return
	}

	value, err := ExtractIdpOrigin(ctx, object)

	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(err.Error()))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, value))
}
