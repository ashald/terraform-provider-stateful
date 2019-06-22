# Terraform Stateful Provider

## Overview

This provider defines generic abstract stateful resources that allow you to manage arbitrary objects by executing
arbitrary commands.

The main principle is that the one can rely on
[external data source](https://www.terraform.io/docs/providers/external/data_source.html) to execute an arbitrary
command to retrieve the real state of the object. And in conjunction with `stateful_*` resource it's possible to invoke
arbitrary [provisioner](https://www.terraform.io/docs/provisioners/index.html) upon resource creation, update or
deletion.

## Resources

This plugin defines following resources:
* `stateful_map` (both keys and values must be strings)
* `stateful_string`

Generally speaking, it should be possible to handle arbitrary configurations with `stateful_string` if object's real
state is handled as an opaque string (for instance generated with 
[jsonencode](https://www.terraform.io/docs/configuration/interpolation.html#jsonencode-value-)). The `stateful_map` 
resource is add as a convenience shortcut for cases when object's state can be described as a JSON map with keys and
values being strings.  

All input arguments and output attributes are the same for all resources.

## Reference

### Arguments

The following arguments are supported:

* `desired` - (Required) State that presumable will be enforced by `provisioner`s upon creation/update,
serves as a trigger for updates. Used for fingerprinting via `hash` attribute (see below).
* `real` - (Optional) An optional feedback about the "real" state of the object. When set, allows Terraform to detect
situations when real state diverges from the desired one (for instance, an update outside of Terraform configuration).  

All arguments must be of the same type and depend on the resource:
* `string` for `stateful_string` 
* `map[string,string]` for `stateful_map`

### Attributes

The following attribute is exported:

* `hash` - The "fingerprint" of the `desired` state of the resource that can be used with
[null_resource](https://www.terraform.io/docs/providers/null/resource.html)'s `triggers` argument in order to invoke
update actions. Currently SHA256 of the JSON representation of `desired` argument is used. 

## Limitations

### No meaningful diffs for `real` argument

Due to limitations of Terraform API, there is no [feasible] way to display meaningful diffs for `real` attribute in case
when object diverges from Terraform configuration. In order to reduce confusion and maintain uniform behavior `real`
field's diffs are always rendered as:
```
real.%: "" => <computed>
```

### Destroy Provisioners

Due to limitations in current implementation of destroy provisioners they are not executed when resource definition is
removed from Terraform configuration. Instead `count` meta-parameter should be used. See
[official documentation](https://www.terraform.io/docs/provisioners/index.html#destroy-time-provisioners) for details.


## Installation

> Terraform automatically discovers the Providers when it parses configuration files.
> This only occurs when the init command is executed.

Currently Terraform is able to automatically download only [official plugins distributed by HashiCorp](https://github.com/terraform-providers).

The provider plugin can be installed automatically via [Para - 3rd-party plugin manager for Terraform](https://github.com/paraterraform/para)
or it can be downloaded and installed manually.  

### Para

This plugin is available via [default index](https://github.com/paraterraform/index) for [Para](https://github.com/paraterraform/para).
If you use Para or Para Launcher you can just skip to the [Usage](#usage) section below assuming you'd wrap all calls to Terraform with Para:
```bash
$ ./para terraform init
Para Launcher Activated!
- Checking para.cfg.yaml in current directory for 'version: X.Y.Z'
- Desired version: latest (latest is used when no version specified)
- Executing '$TMPDIR/para-501/para/latest/para_v0.3.1_darwin-amd64'

------------------------------------------------------------------------

Para is being initialized...
- Cache Dir: $TMPDIR/para-501
- Terraform: downloading to $TMPDIR/para-501/terraform/0.12.2/darwin_amd64
- Plugin Dir: terraform.d/plugins
- Primary Index: https://raw.githubusercontent.com/paraterraform/index/master/para.idx.yaml as of 2019-06-22T00:59:24-04:00 (providers: 16)
- Index Extensions: para.idx.d (0/0), ~/.para/para.idx.d (0/0), /etc/para/para.idx.d (0/0)
- Command: terraform init

------------------------------------------------------------------------


Initializing the backend...

Initializing provider plugins...
- Para provides 3rd-party Terraform provider plugin 'stateful' version 'v1.1.0' for 'darwin_amd64' (downloading)


The following providers do not have any version constraints in configuration,
so the latest version was installed.

To prevent automatic upgrades to new major versions that may contain breaking
changes, it is recommended to add version = "..." constraints to the
corresponding provider blocks in configuration, with the constraint strings
suggested below.

* provider.stateful: version = "~> 1.1"

Terraform has been successfully initialized!
```  

If you use Para but don't use the [default index](https://github.com/paraterraform/index) you can make the plugin available
by including index extension for this plugin: either add [`provider.stateful.yaml`](./provider.stateful.yaml) from this
repo to your [Para index extensions dir](https://github.com/paraterraform/para#extensions) to fix currently available versions
or create `provider.stateful.yaml` as an empty file and put the URL to the aforementioned file inside to automatically get updates:
```yaml
https://raw.githubusercontent.com/ashald/terraform-provider-stateful/master/go.mod
```

### Manual

> Terraform will search for matching Providers via a
> [Discovery](https://www.terraform.io/docs/extend/how-terraform-works.html#discovery) process, **including the current
> local directory**.

This means that the plugin should either be placed into current working directory where Terraform will be executed from
or it can be [installed system-wide](https://www.terraform.io/docs/configuration/providers.html#third-party-plugins).

## Usage

### main.tf
```hcl
locals {
  desired = "desired.json"
  real    = "real.json"
}

data "external" "desired" {
  program = ["cat", local.desired]
}

data "external" "real" {
  program = ["cat", local.real]
}

resource "stateful_map" "my_resource" {
  // The "count" meta-parameter is used to address destroy provisioner limitation
  // See https://www.terraform.io/docs/provisioners/index.html#destroy-time-provisioners for details
  // For the sake for usage example we read value from file, in real world set it explicitely
  count = trimspace(file("count"))

  desired = data.external.desired.result
  real    = data.external.real.result

  provisioner "local-exec" {
    command = format("echo '%s' > %s", jsonencode(self.desired), local.real)
  }
  provisioner "local-exec" {
    when    = destroy
    command = format("echo {} > %s", local.real)
  }
}

resource "null_resource" "updates" {
  count = trimspace(file("count"))

  triggers = {
    state = stateful_map.my_resource[count.index].hash
  }

  provisioner "local-exec" {
    command = format("echo '%s' > %s", jsonencode(stateful_map.my_resource[count.index].desired), local.real)
  }
}

```

### Download
```bash
wget "https://github.com/ashald/terraform-provider-stateful/releases/download/v1.1.0/terraform-provider-stateful_v1.1.0-$(uname -s | tr '[:upper:]' '[:lower:]')-amd64"
chmod +x ./terraform-provider-stateful*
```

### Init
```bash
$ ls -1
  main.tf
  terraform-provider-stateful_v1.1.0-linux-amd64

$ terraform init
  
  Initializing the backend...
  
  Initializing provider plugins...
  - Checking for available provider plugins...
  - Downloading plugin for provider "external" (terraform-providers/external) 1.1.2...
  - Downloading plugin for provider "null" (terraform-providers/null) 2.1.2...
  
  The following providers do not have any version constraints in configuration,
  so the latest version was installed.
  
  To prevent automatic upgrades to new major versions that may contain breaking
  changes, it is recommended to add version = "..." constraints to the
  corresponding provider blocks in configuration, with the constraint strings
  suggested below.
  
  * provider.external: version = "~> 1.1"
  * provider.null: version = "~> 2.1"
  * provider.stateful: version = "~> 1.0"
  
  Terraform has been successfully initialized!
  
  You may now begin working with Terraform. Try running "terraform plan" to see
  any changes that are required for your infrastructure. All Terraform commands
  should now work.
  
  If you ever set or change modules or backend configuration for Terraform,
  rerun this command to reinitialize your working directory. If you forget, other
  commands will detect it and remind you to do so if necessary.
```

### Create
```bash
$ echo '{"foo":"bar"}' > desired.json

# The `external` data source used in this example requires JSON input
# For the sake of simplicity in this example we read `real.json` using `cat`
# So we have to initialize the file with an empty JSON object 
$ echo '{}' > real.json

# The "count" meta-parameter is used to address destroy provisioner limitation
# See https://www.terraform.io/docs/provisioners/index.html#destroy-time-provisioners for details
$ echo 1 > count
 

$ terraform apply
  data.external.real: Refreshing state...
  data.external.desired: Refreshing state...
  
  An execution plan has been generated and is shown below.
  Resource actions are indicated with the following symbols:
    + create
  
  Terraform will perform the following actions:
  
    # null_resource.updates[0] will be created
    + resource "null_resource" "updates" {
        + id       = (known after apply)
        + triggers = (known after apply)
      }
  
    # stateful_map.my_resource[0] will be created
    + resource "stateful_map" "my_resource" {
        + desired = {
            + "foo" = "bar"
          }
        + hash    = (known after apply)
        + id      = (known after apply)
        + real    = (known after apply)
      }
  
  Plan: 2 to add, 0 to change, 0 to destroy.
  
  Do you want to perform these actions?
    Terraform will perform the actions described above.
    Only 'yes' will be accepted to approve.
  
    Enter a value: yes
  
  stateful_map.my_resource[0]: Creating...
  stateful_map.my_resource[0]: Provisioning with 'local-exec'...
  stateful_map.my_resource[0] (local-exec): Executing: ["/bin/sh" "-c" "echo '{\"foo\":\"bar\"}' > real.json"]
  stateful_map.my_resource[0]: Creation complete after 0s [id=05f5e31d-6b5b-41e8-b15d-6a6774111598]
  null_resource.updates[0]: Creating...
  null_resource.updates[0]: Provisioning with 'local-exec'...
  null_resource.updates[0] (local-exec): Executing: ["/bin/sh" "-c" "echo '{\"foo\":\"bar\"}' > real.json"]
  null_resource.updates[0]: Creation complete after 0s [id=5046540171915034813]
  
  Apply complete! Resources: 2 added, 0 changed, 0 destroyed.

$ cat real.json
  {"foo":"bar"}
 
```
### Update

```bash
$ echo '{"foo":"baz"}' > desired.json

$ terraform apply
  data.external.real: Refreshing state...
  data.external.desired: Refreshing state...
  stateful_map.my_resource[0]: Refreshing state... [id=05f5e31d-6b5b-41e8-b15d-6a6774111598]
  null_resource.updates[0]: Refreshing state... [id=5046540171915034813]
  
  An execution plan has been generated and is shown below.
  Resource actions are indicated with the following symbols:
    ~ update in-place
  -/+ destroy and then create replacement
  
  Terraform will perform the following actions:
  
    # null_resource.updates[0] must be replaced
  -/+ resource "null_resource" "updates" {
        ~ id       = "5046540171915034813" -> (known after apply)
        ~ triggers = {
            - "state" = "7a38bf81f383f69433ad6e900d35b3e2385593f76a7b7ab5d4355b8ba41ee24b"
          } -> (known after apply) # forces replacement
      }
  
    # stateful_map.my_resource[0] will be updated in-place
    ~ resource "stateful_map" "my_resource" {
        ~ desired = {
            ~ "foo" = "bar" -> "baz"
          }
        ~ hash    = "7a38bf81f383f69433ad6e900d35b3e2385593f76a7b7ab5d4355b8ba41ee24b" -> (known after apply)
          id      = "05f5e31d-6b5b-41e8-b15d-6a6774111598"
        + real    = (known after apply)
      }
  
  Plan: 1 to add, 1 to change, 1 to destroy.
  
  Do you want to perform these actions?
    Terraform will perform the actions described above.
    Only 'yes' will be accepted to approve.
  
    Enter a value: yes
  
  null_resource.updates[0]: Destroying... [id=5046540171915034813]
  null_resource.updates[0]: Destruction complete after 0s
  stateful_map.my_resource[0]: Modifying... [id=05f5e31d-6b5b-41e8-b15d-6a6774111598]
  stateful_map.my_resource[0]: Modifications complete after 0s [id=05f5e31d-6b5b-41e8-b15d-6a6774111598]
  null_resource.updates[0]: Creating...
  null_resource.updates[0]: Provisioning with 'local-exec'...
  null_resource.updates[0] (local-exec): Executing: ["/bin/sh" "-c" "echo '{\"foo\":\"baz\"}' > real.json"]
  null_resource.updates[0]: Creation complete after 0s [id=1242536504768383134]
  
  Apply complete! Resources: 1 added, 1 changed, 1 destroyed.

$ cat real.json
  {"foo":"baz"}

```

### Reconcile

```bash
$ echo '{"foo":"wrong"}' > real.json # diverge the real state

$ terraform apply
  data.external.desired: Refreshing state...
  data.external.real: Refreshing state...
  stateful_map.my_resource[0]: Refreshing state... [id=05f5e31d-6b5b-41e8-b15d-6a6774111598]
  null_resource.updates[0]: Refreshing state... [id=1242536504768383134]
  
  An execution plan has been generated and is shown below.
  Resource actions are indicated with the following symbols:
    ~ update in-place
  -/+ destroy and then create replacement
  
  Terraform will perform the following actions:
  
    # null_resource.updates[0] must be replaced
  -/+ resource "null_resource" "updates" {
        ~ id       = "1242536504768383134" -> (known after apply)
        ~ triggers = {
            - "state" = "c450c726579d41e1daa46158c07c1ed4a81dddc5e8dcb96ad729bca95e0e6fac"
          } -> (known after apply) # forces replacement
      }
  
    # stateful_map.my_resource[0] will be updated in-place
    ~ resource "stateful_map" "my_resource" {
          desired = {
              "foo" = "baz"
          }
        ~ hash    = "c450c726579d41e1daa46158c07c1ed4a81dddc5e8dcb96ad729bca95e0e6fac" -> (known after apply)
          id      = "05f5e31d-6b5b-41e8-b15d-6a6774111598"
        + real    = (known after apply)
      }
  
  Plan: 1 to add, 1 to change, 1 to destroy.
  
  Do you want to perform these actions?
    Terraform will perform the actions described above.
    Only 'yes' will be accepted to approve.
  
    Enter a value: yes
  
  null_resource.updates[0]: Destroying... [id=1242536504768383134]
  null_resource.updates[0]: Destruction complete after 0s
  stateful_map.my_resource[0]: Modifying... [id=05f5e31d-6b5b-41e8-b15d-6a6774111598]
  stateful_map.my_resource[0]: Modifications complete after 0s [id=05f5e31d-6b5b-41e8-b15d-6a6774111598]
  null_resource.updates[0]: Creating...
  null_resource.updates[0]: Provisioning with 'local-exec'...
  null_resource.updates[0] (local-exec): Executing: ["/bin/sh" "-c" "echo '{\"foo\":\"baz\"}' > real.json"]
  null_resource.updates[0]: Creation complete after 0s [id=835260447911403434]
  
  Apply complete! Resources: 1 added, 1 changed, 1 destroyed.

```

### Delete
```bash
# The "count" meta-parameter is used to address destroy provisioner limitation
# See https://www.terraform.io/docs/provisioners/index.html#destroy-time-provisioners for details
$ echo 0 > count 

$ terraform apply
  data.external.desired: Refreshing state...
  data.external.real: Refreshing state...
  stateful_map.my_resource[0]: Refreshing state... [id=05f5e31d-6b5b-41e8-b15d-6a6774111598]
  null_resource.updates[0]: Refreshing state... [id=835260447911403434]
  
  An execution plan has been generated and is shown below.
  Resource actions are indicated with the following symbols:
    - destroy
  
  Terraform will perform the following actions:
  
    # null_resource.updates[0] will be destroyed
    - resource "null_resource" "updates" {
        - id       = "835260447911403434" -> null
        - triggers = {
            - "state" = "c450c726579d41e1daa46158c07c1ed4a81dddc5e8dcb96ad729bca95e0e6fac"
          } -> null
      }
  
    # stateful_map.my_resource[0] will be destroyed
    - resource "stateful_map" "my_resource" {
        - desired = {
            - "foo" = "baz"
          } -> null
        - hash    = "c450c726579d41e1daa46158c07c1ed4a81dddc5e8dcb96ad729bca95e0e6fac" -> null
        - id      = "05f5e31d-6b5b-41e8-b15d-6a6774111598" -> null
      }
  
  Plan: 0 to add, 0 to change, 2 to destroy.
  
  Do you want to perform these actions?
    Terraform will perform the actions described above.
    Only 'yes' will be accepted to approve.
  
    Enter a value: yes
  
  null_resource.updates[0]: Destroying... [id=835260447911403434]
  null_resource.updates[0]: Destruction complete after 0s
  stateful_map.my_resource[0]: Destroying... [id=05f5e31d-6b5b-41e8-b15d-6a6774111598]
  stateful_map.my_resource[0]: Provisioning with 'local-exec'...
  stateful_map.my_resource[0] (local-exec): Executing: ["/bin/sh" "-c" "echo {} > real.json"]
  stateful_map.my_resource[0]: Destruction complete after 0s
  
  Apply complete! Resources: 0 added, 0 changed, 2 destroyed.

```

## Development

## Go

In order to work on the provider, [Go](http://www.golang.org) should be installed first (version 1.11+ is *required*).
[goenv](https://github.com/syndbg/goenv) and [gvm](https://github.com/moovweb/gvm) are great utilities that can help a
lot with that and simplify setup tremendously. 
[GOPATH](http://golang.org/doc/code.html#GOPATH) should be setup correctly and `$GOPATH/bin` should be
added `$PATH`.

This plugin uses Go modules available starting from Go `1.11` and therefore it **should not** be checked out within `$GOPATH` tree.

## Source Code

Source code can be retrieved with `git`
```bash
$ git clone git@github.com:ashald/terraform-provider-stateful.git .
```

## Dependencies

This project uses `go mod` to manage its dependencies and it's expected that all dependencies are vendored so that
it's buildable without internet access. When adding/removing a dependency run following commands:
```bash
$ go mod vendor
$ go mod tidy
```

### Test

```bash
$ make clean format test
  rm -rf ./release terraform-provider-stateful_v1.1.0
  go fmt ./...
  go test -v ./...
  ?   	github.com/ashald/terraform-provider-stateful	[no test files]
  === RUN   TestProvider
  --- PASS: TestProvider (0.00s)
  === RUN   TestStatefulString
  --- PASS: TestStatefulString (0.12s)
  PASS
  ok  	github.com/ashald/terraform-provider-stateful/stateful	(cached)
  go vet ./...
```

### Build
In order to build plugin for the current platform use [GNU]make:
```bash
$ make build
  go build -o terraform-provider-stateful_v1.1.0

```

it will build provider from sources and put it into current working directory.

If Terraform was installed (as a binary) or via `go get -u github.com/hashicorp/terraform` it'll pick up the plugin if 
executed against a configuration in the same directory.

### Release

In order to prepare provider binaries for all platforms:
```bash
$ make release
  GOOS=darwin GOARCH=amd64 go build -o './release/terraform-provider-stateful_v1.1.0-darwin-amd64'
  GOOS=linux GOARCH=amd64 go build -o './release/terraform-provider-stateful_v1.1.0-linux-amd64'
```

### Versioning

This project follow [Semantic Versioning](https://semver.org/)

### Changelog

This project follows [keep a changelog](https://keepachangelog.com/en/1.0.0/) guidelines for changelog.

### Contributors

Please see [CONTRIBUTORS.md](./CONTRIBUTORS.md)

## License

This is free and unencumbered software released into the public domain. See [LICENSE](./LICENSE)
