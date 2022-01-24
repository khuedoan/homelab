# Try on a VM

## Prerequisites

Install the following packages:

- `docker`
- `make`
- `vagrant`
- `virtualbox`

VM specifications:

```ruby
# metal/Vagrantfile
{{#include ../../metal/Vagrantfile:vm_specs}}
```

## Configuration

Follow the the steps in the [configuration guide](./deployment/configuration.md).

## Build

```sh
VAGRANT_CWD=./metal vagrant up
```

Open the tools container:

```sh
make tools
```

Build a cluster on the test VM and bootstrap it:

```
make dev
```

Finally follow the [DNS guide](./deployment/dns.md) to update your DNS setup.

Now you can visit the dashboard at <https://home.example.com>

## Clean up

Shut down the VM:

```sh
cd metal
vagrant destroy
```

Then clean up your DNS config from the previous step.
