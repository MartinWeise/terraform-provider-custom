# terraform import btp_directory.<resource_name> <display_id>,<tenant_subdomain>

terraform import custom_business_partner_company_group_contact.main SEQ-1,porr-dev

# terraform import using id attribute in import block

import {
  to = custom_business_partner_company_group_contact.<resource_name>
  id = "<display_id>,<tenant_subdomain>"
}

import {
  to = custom_business_partner_company_group_contact.<resource_name>
  identity = {
    directory_id = "<display_id>,<tenant_subdomain>"
  }
}