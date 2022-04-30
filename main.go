package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"terraform-provider-dependencies/internal/provider"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: provider.New})
}
