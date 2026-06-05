# This will return a string value
output "idp_origin" {
  value = provider::custom::extract_idp_origin(data.custom_destination_saml_metadata.subaccount_gateway.idp_metadata)
  sensitive   = true
  description = "SAML 2.0 IdP Origin"
}