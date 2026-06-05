# This will return a string value
output "cis_url" {
  value = provider::custom::extract_identity_service_url(btp_subaccount_subscription.identity_services.subscription_url)
  sensitive   = true
  description = "SAP Cloud Identity Service API URL"
}