package main

import (
	"github.com/asteris-llc/terraform-provisioner-converge/converge"

	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
)

const Name = "terraform-provisioner-converge"
const Version = "0.3.0"

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProvisionerFunc: func() terraform.ResourceProvisioner {
			return new(converge.ResourceProvisioner)
		},
	})
}
