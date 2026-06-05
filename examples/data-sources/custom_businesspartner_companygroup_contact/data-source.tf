# read a company group contact
data "custom_businesspartner_companygroup_contact" "main" {
  display_id       = "SEQ-1"
  tenant_subdomain = "liefermannag-dev"
}
