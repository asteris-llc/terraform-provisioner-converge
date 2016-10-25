# terraform-provisioner-converge

Terraform provisioner for Converge

This plugin allows you to run converge directly in terraform.

## Installing

Download a binary from the [releases page](https://github.com/asteris-llc/terraform-provisioner-converge/releases).

Then create a `.terraformrc` file in your home directory with the `converge` provisioner 
pointing to the location of your downloaded binary.

```hcl
provisioners {
  converge = "/usr/local/bin/terraform-provisioner-converge"
}
```


## Building

```shell
$ make vendor
$ make
```

## Options

* `hcl` (required) - A list of files to run through Converge. Can be a single path as a string
* `ca_file` (optional) - Path to a CA certificate to trust. Requires `use_ssl`
* `cert_file` (optional) - Path to a certificate file for SSL. Requires `use_ssl`
* `download_binary` (optional) - Install Converge binary. Default to `false`
* `key_file` (optional) - Path to a key file for SSL. Requires `use_ssl`
* `local` (optional) - Run Converge in local mode. Defaults to `true`
* `local_addr` (optional) - Address to use for local RPC connection
* `log_level` (optional) - Logging level. One of `DEBUG`, `INFO`, `WARN`, `ERROR` or `FATAL`. Default is `INFO`
* `no_token` (optional) - Don't use or generate an RPC token
* `params` (optional) - A hash of parameter/value pairs to pass to Converge
* `rpc_addr` (optional) - Address for server RPC connection. Overrides `local`
* `rpc_token` (optional) - Token to use for RPC
* `use_ssl` (optional) - Use SSL to connect
* `version` (optional) - Specify the version of Converge. Default is the latest version.

## Example Usage

```
resource "aws_instance" "web" {
  ...
  provisioner "converge" {
    params = {
      key = "value"
      message = <<EOF
This is not the default message
EOF
    }
    hcl = [
      "http://some.webserver.com/converge/app.hcl",
      "http://some.other.webserver.org/converge/otherapp.hcl"
    ]
    download_binary = true
    prevent_sudo = false
    # http_proxy = "Outgoing http proxy address"
    # https_proxy = "Outgoing https proxy address"
    # no_proxy = [ "list of ip addresses to not proxy" ]
  }
}
```
