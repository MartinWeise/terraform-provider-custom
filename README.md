[![Release](https://github.com/MartinWeise/terraform-provider-custom/actions/workflows/release.yml/badge.svg)](https://github.com/MartinWeise/terraform-provider-custom/actions/workflows/release.yml)

tl;dr: custom provider to automate SAP BTP with Terraform

## Features

* Manage Credential Store Passwords
* Manage Business Partners
* Read SAML 2.0 Destination Trust Certificates

## Usage

```terraform
terraform {
  required_providers {
    custom = {
      source  = "MartinWeise/custom"
      version = "1.0.3"
    }
  }
}

provider "custom" {
  credstore_binding_parameters   = "{}"
  destination_binding_parameters = "{}"
}
```

## Further Documentation

* [HashiCorp Registry Documentation](https://registry.terraform.io/providers/MartinWeise/custom/latest/docs)