package converge

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/communicator"
	"github.com/hashicorp/terraform/terraform"
)

const (
	installURL = "https://bintray.com/chrisaubuchon/generic/download_file?file_path=converge"
	binaryPath = "/usr/bin"
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

	// First download the install.sh script from Chef
	err := p.runCommand(o, comm, fmt.Sprintf("%scurl -L -o %s/converge %s", prefix, binaryPath, installURL))
	if err != nil {
		return err
	}

	if err := p.runCommand(o, comm, "chmod 0755 /usr/bin/converge"); err != nil {
		return err
	}

	return nil
}
