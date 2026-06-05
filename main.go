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
		Address:         "terraform.local/sequello/custom",
		Debug:           debug,
		ProtocolVersion: 6,
	})

	if err != nil {
		log.Fatal(err)
	}
}
