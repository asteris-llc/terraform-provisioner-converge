package converge

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/communicator"
	"github.com/hashicorp/terraform/terraform"
)

const (
	binaryDir = "/usr/bin"
)

func (p *Provisioner) installConvergeBinary(
	o terraform.UIOutput,
	comm communicator.Communicator) error {

	// Build up the command prefix
	prefix := ""
	if p.HTTPProxy != "" {
		prefix += fmt.Sprintf("http_proxy='%s' ", p.HTTPProxy)
	}
	if p.HTTPSProxy != "" {
		prefix += fmt.Sprintf("https_proxy='%s' ", p.HTTPSProxy)
	}
	if p.NOProxy != nil {
		prefix += fmt.Sprintf("no_proxy='%s' ", strings.Join(p.NOProxy, ","))
	}

	err := p.runCommand(o, comm, fmt.Sprintf("%scurl -LO %s", prefix, installURL))
	if err != nil {
		return err
	}

	err = p.runCommand(o, comm, fmt.Sprintf("%ssh ./install-converge.sh -v %q -d %q", prefix, p.Version, p.InstallDir))
	if err != nil {
		return err
	}

	return p.runCommand(o, comm, fmt.Sprintf("%srm -f install-converge.sh", prefix))
}
