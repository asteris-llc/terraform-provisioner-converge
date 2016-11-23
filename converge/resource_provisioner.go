package converge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform/communicator"
	"github.com/hashicorp/terraform/communicator/remote"
	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/go-linereader"
	"github.com/mitchellh/mapstructure"
)

const (
	installURL = "https://get.converge.sh"
)

type Provisioner struct {
	Download   bool                   `mapstructure:"download_binary"`
	CaFile     string                 `mapstructure:"ca_file"`
	CertFile   string                 `mapstructure:"cert_file"`
	KeyFile    string                 `mapstructure:"key_file"`
	Local      bool                   `mapstructure:"local"`
	LocalAddr  string                 `mapstructure:"local_addr"`
	LogLevel   string                 `mapstructure:"log_level"`
	Hcl        []string               `mapstructure:"hcl"`
	NoToken    bool                   `mapstructure:"no_token"`
	Params     map[string]interface{} `mapstructure:"params"`
	RpcAddr    string                 `mapstructure:"rpc_addr"`
	RpcToken   string                 `mapstructure:"rpc_token"`
	UseSsl     bool                   `mapstructure:"use_ssl"`
	Version    string                 `mapstructure:"version"`
	BinaryDir  string                 `mapstructure:"binary_dir"`
	InstallDir string                 `mapstructure:"install_dir"`

	HTTPProxy   string   `mapstructure:"http_proxy"`
	HTTPSProxy  string   `mapstructure:"https_proxy"`
	NOProxy     []string `mapstructure:"no_proxy"`
	PreventSudo bool     `mapstructure:"prevent_sudo"`

	useSudo bool
}

type ResourceProvisioner struct{}

func (r *ResourceProvisioner) Apply(
	o terraform.UIOutput,
	s *terraform.InstanceState,
	c *terraform.ResourceConfig) error {

	p, err := r.decodeConfig(c)
	if err != nil {
		return err
	}

	p.useSudo = !p.PreventSudo && s.Ephemeral.ConnInfo["user"] != "root"

	if p.BinaryDir == "" {
		p.BinaryDir = binaryDir
	}

	// Get a new communicator
	comm, err := communicator.New(s)
	if err != nil {
		return err
	}

	// Wait and retry until we establish the connection
	err = retryFunc(comm.Timeout(), func() error {
		err := comm.Connect(o)
		return err
	})
	if err != nil {
		return err
	}
	defer comm.Disconnect()

	if p.Download {
		o.Output("Installing converge...")
		if err := p.installConvergeBinary(o, comm); err != nil {
			return err
		}
	}

	o.Output("Running converge...")
	if err := p.runConverge(o, comm); err != nil {
		return err
	}

	return nil
}

func (r *ResourceProvisioner) Validate(c *terraform.ResourceConfig) (ws []string, es []error) {
	p, err := r.decodeConfig(c)
	if err != nil {
		es = append(es, err)
		return ws, es
	}

	if len(p.Hcl) == 0 {
		es = append(es, fmt.Errorf("No modules selected"))
	}

	return ws, es
}

func (r *ResourceProvisioner) decodeConfig(c *terraform.ResourceConfig) (*Provisioner, error) {
	p := new(Provisioner)

	// Set plugin defaults here
	p.Download = true

	decConf := &mapstructure.DecoderConfig{
		ErrorUnused:      true,
		WeaklyTypedInput: true,
		Result:           p,
	}

	dec, err := mapstructure.NewDecoder(decConf)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	for k, v := range c.Raw {
		m[k] = v
	}

	for k, v := range c.Config {
		m[k] = v
	}

	if err := dec.Decode(m); err != nil {
		return nil, err
	}

	return p, nil
}

// runCommand is used to run already prepared commands
func (p *Provisioner) runCommand(
	o terraform.UIOutput,
	comm communicator.Communicator,
	command string) error {
	var err error

	if p.useSudo {
		command = "sudo " + command
	}

	outR, outW := io.Pipe()
	errR, errW := io.Pipe()
	outDoneCh := make(chan struct{})
	errDoneCh := make(chan struct{})
	go p.copyOutput(o, outR, outDoneCh)
	go p.copyOutput(o, errR, errDoneCh)

	cmd := &remote.Cmd{
		Command: command,
		Stdout:  outW,
		Stderr:  errW,
	}

	if err := comm.Start(cmd); err != nil {
		return fmt.Errorf("Error executing command %q: %v", cmd.Command, err)
	}

	cmd.Wait()
	if cmd.ExitStatus != 0 {
		err = fmt.Errorf(
			"Command %q exited with non-zero exit status: %d", cmd.Command, cmd.ExitStatus)
	}

	// Wait for output to clean up
	outW.Close()
	errW.Close()
	<-outDoneCh
	<-errDoneCh

	// If we have an error, return it out now that we've cleaned up
	if err != nil {
		return err
	}

	return nil
}

// retryFunc is used to retry a function for a given duration
func retryFunc(timeout time.Duration, f func() error) error {
	finish := time.After(timeout)
	for {
		err := f()
		if err == nil {
			return nil
		}
		log.Printf("Retryable error: %v", err)

		select {
		case <-finish:
			return err
		case <-time.After(3 * time.Second):
		}
	}
}

func (p *Provisioner) copyOutput(o terraform.UIOutput, r io.Reader, doneCh chan<- struct{}) {
	defer close(doneCh)
	lr := linereader.New(r)
	for line := range lr.Ch {
		o.Output(line)
	}
}

func (p *Provisioner) runConverge(o terraform.UIOutput, comm communicator.Communicator) error {
	cmd, err := p.buildCommandLine()
	if err != nil {
		return err
	}

	return p.runCommand(o, comm, cmd)
}

func (p *Provisioner) buildCommandLine() (string, error) {
	cmd := bytes.NewBufferString(fmt.Sprintf("%s/converge apply", p.BinaryDir))

	// An RPC address takes precedence over a local address
	if p.RpcAddr != "" {
		cmd.WriteString(fmt.Sprintf(" --rpc-addr='%s'", p.RpcAddr))
	} else {
		cmd.WriteString(" --local")
		if p.LocalAddr != "" {
			cmd.WriteString(fmt.Sprintf(" --local-addr='%s'", p.LocalAddr))
		}
	}

	if p.NoToken {
		cmd.WriteString(" --no-token")
	}

	if p.RpcToken != "" {
		cmd.WriteString(fmt.Sprintf(" --rpc-token='%s'", p.RpcToken))
	}

	if p.UseSsl {
		cmd.WriteString(" --use-ssl")
		if p.CaFile != "" {
			cmd.WriteString(fmt.Sprintf(" --ca-file='%s'", p.CaFile))
		}
		if p.CertFile != "" {
			cmd.WriteString(fmt.Sprintf(" --cert-file='%s'", p.CertFile))
		}
		if p.KeyFile != "" {
			cmd.WriteString(fmt.Sprintf(" --key-file='%s'", p.KeyFile))
		}
	}

	if p.Params != nil {
		params, err := json.Marshal(p.Params)
		if err != nil {
			return "", err
		}
		cmd.WriteString(" --paramsJSON='")
		cmd.Write(params)
		cmd.WriteString("'")
	}

	cmd.WriteString(" ")
	cmd.WriteString(strings.Join(p.Hcl, " "))

	return cmd.String(), nil
}
