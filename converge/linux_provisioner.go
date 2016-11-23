package converge

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/communicator"
	"github.com/hashicorp/terraform/terraform"
)

const (
	scriptName = "install-converge.sh"
	binaryDir  = "/usr/local/bin"
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

	opts := ""
	if p.Version != "" {
		opts += fmt.Sprintf("-v %s", p.Version)
	}

	if p.InstallDir != "" {
		opts += fmt.Sprintf("-d %s", p.InstallDir)
	}

	err := p.runCommand(o, comm, fmt.Sprintf("%scurl -L %s -o %s", prefix, installURL, scriptName))
	if err != nil {
		return err
	}

	err = p.runCommand(o, comm, fmt.Sprintf("%ssh %s %s", prefix, scriptName, opts))
	if err != nil {
		return err
	}

	return p.runCommand(o, comm, fmt.Sprintf("%srm -f %s", prefix, scriptName))
}
