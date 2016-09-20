package converge

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func test_makeCommand(s string) string {
	return fmt.Sprintf("%s/converge apply %s ", binaryPath, s)
}

func TestBuildCommandLine(t *testing.T) {
	p := new(Provisioner)

	// Basic test. --local is the default so it is expected
	cmd, err := p.buildCommandLine()
	assert.Nil(t, err)
	assert.Equal(t, test_makeCommand("--local"), cmd)

	// Local addr test
	p.LocalAddr = "1.2.3.4:5678"
	cmd, err = p.buildCommandLine()
	assert.Nil(t, err)
	assert.Equal(t, test_makeCommand("--local --local-addr='1.2.3.4:5678'"), cmd)

	// RPC addr test. Leave Local and LocalAddr set to verify that they are not set
	// in the final command line
	p.RpcAddr = "8.7.6.5:4321"
	cmd, err = p.buildCommandLine()
	assert.Nil(t, err)
	assert.Equal(t, test_makeCommand("--rpc-addr='8.7.6.5:4321'"), cmd)

	// Clear RPC addr. --local is expected in the following tests
	p.RpcAddr = ""
	p.LocalAddr = ""
	p.Local = false

	// NoToken test
	p.NoToken = true
	cmd, err = p.buildCommandLine()
	assert.Nil(t, err)
	assert.Equal(t, test_makeCommand("--local --no-token"), cmd)
	p.NoToken = false

	// RpcToken
	p.RpcToken = "1234"
	cmd, err = p.buildCommandLine()
	assert.Nil(t, err)
	assert.Equal(t, test_makeCommand("--local --rpc-token='1234'"), cmd)
	p.RpcToken = ""

	p.UseSsl = false

	// Verify CaFile, CertFile and KeyFile are skipped when UseSsl == false
	p.CaFile = "ca_file"
	p.CertFile = "cert_file"
	p.KeyFile = "key_file"
	cmd, err = p.buildCommandLine()
	assert.Nil(t, err)
	assert.Equal(t, test_makeCommand("--local"), cmd)

	// UseSsl
	p.UseSsl = true
	cmd, err = p.buildCommandLine()
	assert.Nil(t, err)
	assert.Equal(t, test_makeCommand("--local --use-ssl --ca-file='ca_file' --cert-file='cert_file' --key-file='key_file'"), cmd)
	p.UseSsl = false

	// Params
	p.Params = map[string]interface{}{"test": "tset"}
	cmd, err = p.buildCommandLine()
	assert.Nil(t, err)
	assert.Equal(t, test_makeCommand("--local --paramsJSON='{\"test\":\"tset\"}'"), cmd)
	p.Params = nil
}
