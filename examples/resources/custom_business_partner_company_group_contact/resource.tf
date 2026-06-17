resource "custom_business_partner_company_group_contact" "main" {
  display_id       = "SEQ-1"
  tenant_subdomain = "baucharly-dev"
  given_name       = "Foo"
  family_name      = "Bar"
  email_address    = "foo.bar+dev-baucharly@sequello.com"
  phone            = "4212345678"
  mobile           = "4212345678"
}
