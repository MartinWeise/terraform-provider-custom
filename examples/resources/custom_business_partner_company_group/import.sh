# terraform import btp_directory.<resource_name> <display_id>,<subdomain>,<tenant_id>

terraform import custom_business_partner_company_group.main 100000001,porrag-dev,338a52cd-89d6-4daf-8574-daa040bfb700

# terraform import using id attribute in import block

import {
  to = custom_business_partner_company_group.<resource_name>
  id = "<display_id>,<subdomain>,<tenant_id>"
}

import {
  to = custom_business_partner_company_group.<resource_name>
  identity = {
    directory_id = "<display_id>,<subdomain>,<tenant_id>"
  }
}