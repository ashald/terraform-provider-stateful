package main

import (
	"github.com/ashald/terraform-provider-stateful/stateful"
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return stateful.Provider()
		},
	})
}
