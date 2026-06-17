//go:generate go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
//go:generate tfplugindocs generate --rendered-provider-name "SAP BTP Custom"

package main

import (
	"context"
	"flag"
	"log"
	"terraform-provider-custom/btp/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the custom with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address:         "registry.terraform.io/martinweise/custom",
		Debug:           debug,
		ProtocolVersion: 6,
	})

	if err != nil {
		log.Fatal(err)
	}
}
