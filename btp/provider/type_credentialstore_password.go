package provider

import (
	"terraform-provider-custom/btp/client/btp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type credentialStorePasswordType struct {
	Namespace    types.String `tfsdk:"namespace"`
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Value        types.String `tfsdk:"value"`
	Type         types.String `tfsdk:"type"`
	ModifiedAt   types.String `tfsdk:"modified_at"`
	Username     types.String `tfsdk:"username"`
	Metadata     types.String `tfsdk:"metadata"`
	Unmodifiable types.Bool   `tfsdk:"unmodifiable"`
}

func credentialStorePasswordFromValue(obj *btp.Password, namespace string, value string) (credentialStorePasswordType, diag.Diagnostics) {
	var password credentialStorePasswordType

	password.Id = types.StringValue(obj.Id)
	password.Name = types.StringValue(obj.Name)
	password.Type = types.StringValue(obj.Type)
	password.Username = types.StringPointerValue(obj.Username)
	password.ModifiedAt = types.StringPointerValue(obj.ModifiedAt)
	password.Metadata = types.StringPointerValue(obj.Metadata)
	password.Unmodifiable = types.BoolValue(obj.Unmodifiable)
	password.Namespace = types.StringValue(namespace)
	password.Value = types.StringValue(value)

	return password, nil
}
