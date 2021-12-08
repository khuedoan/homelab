# Try on a VM

## Prerequisites

Install the following packages:

- `docker`
- `make`
- `vagrant`
- `virtualbox`

## Build

Change the IP prefix in the Vagrant config to match your LAN setup, for example my IP prefix is `192.168.1.`:

```ruby
# metal/Vagrantfile
{{#include ../../metal/Vagrantfile:4}}
```

Follow the **remaining steps in the [configuration guide](./deployment/configuration.md)**, then create a test VM:

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

Finally follow the [DNS guide](./deployment/dns.md) to update your DNS setup (the easiest one is the nip.io option)

Now you can visit the dashboard at <https://home.example.com>

## Clean up

Shut down the VM:

```sh
cd metal
vagrant destroy -f
```

Then clean up your router config or `/etc/hosts` file from the previous step
