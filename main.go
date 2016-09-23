package main

import (
	"github.com/ChrisAubuchon/terraform-provisioner-converge/converge"

	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProvisionerFunc: func() terraform.ResourceProvisioner {
			return new(converge.ResourceProvisioner)
		},
	})
}
