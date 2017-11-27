package main

import (
	"github.com/hashicorp/terraform/plugin"
//	"github.com/terraform-providers/terraform-provider-hypercloud/hypercloud"
    "github.com/mistarhee/terraform-provider-hypercloud/hypercloud"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: hypercloud.Provider})
}
