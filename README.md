# terraform-provisioner-converge

Terraform provisioner for Converge

## Building

```shell
$ make vendor
$ make
```

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
