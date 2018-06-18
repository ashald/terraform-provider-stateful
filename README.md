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

## Installation

> Terraform automatically discovers the Providers when it parses configuration files.
> This only occurs when the init command is executed.

Currently Terraform is able to automatically download only
[official plugins distributed by HashiCorp](https://github.com/terraform-providers).

[All other plugins](https://www.terraform.io/docs/providers/type/community-index.html) should be installed manually.

> Terraform will search for matching Providers via a
> [Discovery](https://www.terraform.io/docs/extend/how-terraform-works.html#discovery) process, **including the current
> local directory**.

This means that the plugin should either be placed into current working directory where Terraform will be executed from
or it can be [installed system-wide](https://www.terraform.io/docs/configuration/providers.html#third-party-plugins).


## Usage

### main.tf
```hcl
locals { desired="desired.json" real="real.json" }

data "external" "desired" { program=["cat", "${local.desired}"] }
data "external" "real"    { program=["cat", "${local.real}"   ] }

resource "stateful_map" "my_resource" {
  // The "count" meta-parameter is used to address destroy provisioner limitation
  // See https://www.terraform.io/docs/provisioners/index.html#destroy-time-provisioners for details
  // For the sake for usage example we read value from file, in real world set it explicitely
  count = "${trimspace(file("count"))}"
  
  desired = "${data.external.desired.result}"
  real    = "${data.external.real.result}"

  provisioner "local-exec" { command="echo '${jsonencode(stateful_map.my_resource.desired)}' > ${local.real}" }
  provisioner "local-exec" { command="echo {} > ${local.real}" when="destroy"                                 }
}

resource "null_resource" "updates" {
  triggers { state = "${stateful_map.my_resource.hash}" }

  provisioner "local-exec" { command="echo '${jsonencode(stateful_map.my_resource.desired)}' > ${local.real}" }
}
```

### Init
```bash
$ ls -1
  main.tf
  terraform-provider-stateful_v1.0.0

$ terraform init

Initializing provider plugins...

The following providers do not have any version constraints in configuration,
so the latest version was installed.

To prevent automatic upgrades to new major versions that may contain breaking
changes, it is recommended to add version = "..." constraints to the
corresponding provider blocks in configuration, with the constraint strings
suggested below.

* provider.external: version = "~> 1.0"
* provider.null: version = "~> 1.0"
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
  data.external.desired: Refreshing state...
  data.external.real: Refreshing state...
  
  An execution plan has been generated and is shown below.
  Resource actions are indicated with the following symbols:
    + create
  
  Terraform will perform the following actions:
  
    + null_resource.updates
        id:          <computed>
        triggers.%:  <computed>
  
    + stateful_map.my_resource
        id:          <computed>
        desired.%:   "1"
        desired.foo: "bar"
        hash:        <computed>
        real.%:      <computed>
  
  
  Plan: 2 to add, 0 to change, 0 to destroy.
  
  Do you want to perform these actions?
    Terraform will perform the actions described above.
    Only 'yes' will be accepted to approve.
  
    Enter a value: yes
  
  stateful_map.my_resource: Creating...
    desired.%:   "0" => "1"
    desired.foo: "" => "bar"
    hash:        "" => "<computed>"
    real.%:      "" => "<computed>"
  stateful_map.my_resource: Provisioning with 'local-exec'...
  stateful_map.my_resource (local-exec): Executing: ["/bin/sh" "-c" "echo {\"foo\":\"bar\"} > real.json"]
  stateful_map.my_resource: Creation complete after 0s (ID: 40d0a5eb-1a7f-4c6b-a60b-6292baf5d1fd)
  null_resource.updates: Creating...
    triggers.%:     "" => "1"
    triggers.state: "" => "7a38bf81f383f69433ad6e900d35b3e2385593f76a7b7ab5d4355b8ba41ee24b"
  null_resource.updates: Provisioning with 'local-exec'...
  null_resource.updates (local-exec): Executing: ["/bin/sh" "-c" "echo {\"foo\":\"bar\"} > real.json"]
  null_resource.updates: Creation complete after 0s (ID: 6445584234947433393)
  
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
  stateful_map.my_resource: Refreshing state... (ID: 9737bc60-31e6-4d66-b6cc-f12d6b37a29b)
  null_resource.updates: Refreshing state... (ID: 3879522802033916949)
  
  An execution plan has been generated and is shown below.
  Resource actions are indicated with the following symbols:
    ~ update in-place
  -/+ destroy and then create replacement
  
  Terraform will perform the following actions:
  
  -/+ null_resource.updates (new resource required)
        id:             "3879522802033916949" => <computed> (forces new resource)
        triggers.%:     "1" => <computed> (forces new resource)
        triggers.state: "7a38bf81f383f69433ad6e900d35b3e2385593f76a7b7ab5d4355b8ba41ee24b" => "" (forces new resource)
  
    ~ stateful_map.my_resource
        desired.foo:    "bar" => "baz"
        hash:           "7a38bf81f383f69433ad6e900d35b3e2385593f76a7b7ab5d4355b8ba41ee24b" => <computed>
        real.%:         "" => <computed>
  
  
  Plan: 1 to add, 1 to change, 1 to destroy.
  
  Do you want to perform these actions?
    Terraform will perform the actions described above.
    Only 'yes' will be accepted to approve.
  
    Enter a value: yes
  
  null_resource.updates: Destroying... (ID: 3879522802033916949)
  stateful_map.my_resource: Modifying... (ID: 9737bc60-31e6-4d66-b6cc-f12d6b37a29b)
    desired.foo: "bar" => "baz"
    hash:        "7a38bf81f383f69433ad6e900d35b3e2385593f76a7b7ab5d4355b8ba41ee24b" => "<computed>"
    real.%:      "" => "<computed>"
  null_resource.updates: Destruction complete after 0s
  stateful_map.my_resource: Modifications complete after 0s (ID: 9737bc60-31e6-4d66-b6cc-f12d6b37a29b)
  null_resource.updates: Creating...
    triggers.%:     "" => "1"
    triggers.state: "" => "c450c726579d41e1daa46158c07c1ed4a81dddc5e8dcb96ad729bca95e0e6fac"
  null_resource.updates: Provisioning with 'local-exec'...
  null_resource.updates (local-exec): Executing: ["/bin/sh" "-c" "echo '{\"foo\":\"baz\"}' > real.json"]
  null_resource.updates: Creation complete after 0s (ID: 131549942545567523)
  
  Apply complete! Resources: 1 added, 1 changed, 1 destroyed.

$ cat real.json
  {"foo":"baz"}

```

### Reconcile

```bash
$ echo '{"foo":"wrong"}' > real.json # diverge the real state

$ terraform apply
  data.external.real: Refreshing state...
  data.external.desired: Refreshing state...
  stateful_map.my_resource: Refreshing state... (ID: 9737bc60-31e6-4d66-b6cc-f12d6b37a29b)
  null_resource.updates: Refreshing state... (ID: 131549942545567523)
  
  An execution plan has been generated and is shown below.
  Resource actions are indicated with the following symbols:
    ~ update in-place
  -/+ destroy and then create replacement
  
  Terraform will perform the following actions:
  
  -/+ null_resource.updates (new resource required)
        id:             "131549942545567523" => <computed> (forces new resource)
        triggers.%:     "1" => <computed> (forces new resource)
        triggers.state: "c450c726579d41e1daa46158c07c1ed4a81dddc5e8dcb96ad729bca95e0e6fac" => "" (forces new resource)
  
    ~ stateful_map.my_resource
        hash:           "c450c726579d41e1daa46158c07c1ed4a81dddc5e8dcb96ad729bca95e0e6fac" => <computed>
        real.%:         "" => <computed>
  
  
  Plan: 1 to add, 1 to change, 1 to destroy.
  
  Do you want to perform these actions?
    Terraform will perform the actions described above.
    Only 'yes' will be accepted to approve.
  
    Enter a value: yes
  
  null_resource.updates: Destroying... (ID: 131549942545567523)
  null_resource.updates: Destruction complete after 0s
  stateful_map.my_resource: Modifying... (ID: 9737bc60-31e6-4d66-b6cc-f12d6b37a29b)
    hash:   "c450c726579d41e1daa46158c07c1ed4a81dddc5e8dcb96ad729bca95e0e6fac" => "<computed>"
    real.%: "" => "<computed>"
  stateful_map.my_resource: Modifications complete after 0s (ID: 9737bc60-31e6-4d66-b6cc-f12d6b37a29b)
  null_resource.updates: Creating...
    triggers.%:     "" => "1"
    triggers.state: "" => "c450c726579d41e1daa46158c07c1ed4a81dddc5e8dcb96ad729bca95e0e6fac"
  null_resource.updates: Provisioning with 'local-exec'...
  null_resource.updates (local-exec): Executing: ["/bin/sh" "-c" "echo '{\"foo\":\"baz\"}' > real.json"]
  null_resource.updates: Creation complete after 0s (ID: 5458577328069789046)
  
  Apply complete! Resources: 1 added, 1 changed, 1 destroyed.

```

### Delete
```bash
# The "count" meta-parameter is used to address destroy provisioner limitation
# See https://www.terraform.io/docs/provisioners/index.html#destroy-time-provisioners for details
$ echo 0 > count 

$ terraform apply
  data.external.real: Refreshing state...
  data.external.desired: Refreshing state...
  stateful_map.my_resource: Refreshing state... (ID: 626d67ee-cf46-4f19-9cfe-1ec2e45fcafe)
  null_resource.updates: Refreshing state... (ID: 3743826101948009555)
  
  An execution plan has been generated and is shown below.
  Resource actions are indicated with the following symbols:
    - destroy
  
  Terraform will perform the following actions:
  
    - stateful_map.my_resource
  
  
  Plan: 0 to add, 0 to change, 1 to destroy.
  
  Do you want to perform these actions?
    Terraform will perform the actions described above.
    Only 'yes' will be accepted to approve.
  
    Enter a value: yes
  
  stateful_map.my_resource: Destroying... (ID: 626d67ee-cf46-4f19-9cfe-1ec2e45fcafe)
  stateful_map.my_resource: Provisioning with 'local-exec'...
  stateful_map.my_resource (local-exec): Executing: ["/bin/sh" "-c" "echo {} > real.json"]
  stateful_map.my_resource: Destruction complete after 0s
  
  Apply complete! Resources: 0 added, 0 changed, 1 destroyed.
  
  
$ cat real.json
  {}

$ echo > main.tf

$ terraform apply
  null_resource.updates: Refreshing state... (ID: 3743826101948009555)
  
  An execution plan has been generated and is shown below.
  Resource actions are indicated with the following symbols:
    - destroy
  
  Terraform will perform the following actions:
  
    - null_resource.updates
  
  
  Plan: 0 to add, 0 to change, 1 to destroy.
  
  Do you want to perform these actions?
    Terraform will perform the actions described above.
    Only 'yes' will be accepted to approve.
  
    Enter a value: yes
  
  null_resource.updates: Destroying... (ID: 3743826101948009555)
  null_resource.updates: Destruction complete after 0s
  
  Apply complete! Resources: 0 added, 0 changed, 1 destroyed.

```

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

## Development

### Go

In order to work on the provider, [Go](http://www.golang.org) should be installed first (version 1.8+ is *required*).
[goenv](https://github.com/syndbg/goenv) and [gvm](https://github.com/moovweb/gvm) are great utilities that can help a
lot with that and simplify setup tremendously. 
[GOPATH](http://golang.org/doc/code.html#GOPATH) should be setup correctly and as long as `$GOPATH/bin` should be
added `$PATH`.

### Source Code

Source code can be retrieved either with `go get`

```bash
$ go get -u -d github.com/ashald/terraform-provider-stateful
```

or with `git`
```bash
$ mkdir -p ${GOPATH}/src/github.com/ashald/terraform-provider-stateful
$ cd ${GOPATH}/src/github.com/ashald/terraform-provider-stateful
$ git clone git@github.com:ashald/terraform-provider-stateful.git .
```

### Test

```bash
$ make clean format test
  rm -rf ./release terraform-provider-stateful_v1.0.0
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
  go build -o terraform-provider-stateful_v1.0.0

```

it will build provider from sources and put it into current working directory.

If Terraform was installed (as a binary) or via `go get -u github.com/hashicorp/terraform` it'll pick up the plugin if 
executed against a configuration in the same directory.

### Release

In order to prepare provider binaries for all platforms:
```bash
$ make release
  GOOS=darwin GOARCH=amd64 go build -o './release/terraform-provider-stateful_v1.0.0-darwin-amd64'
  GOOS=linux GOARCH=amd64 go build -o './release/terraform-provider-stateful_v1.0.0-linux-amd64'
```
