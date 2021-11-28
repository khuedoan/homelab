# Try on a VM

## Prerequisites

Install the following packages:

- `virtualbox`
- `vagrant`

## Build

Change the IP prefix in the Vagrant config to match your LAN setup, for example my IP prefix is `192.168.1.`:

```ruby
# metal/Vagrantfile
{{#include ../../metal/Vagrantfile:4}}
```

Follow the remaining steps in the [configuration guide](./deployment/configuration.md), then build the cluster and bootstrap it:

```sh
make -C metal dev
make bootstrap
```

Finally follow the [DNS guide](./deployment/dns.md) to update your DNS setup (the easiest one is the `/etc/hosts` option)

## Clean up

Shut down the VM:

```sh
cd metal
vagrant destroy -f
```

Then clean up your router config or `/etc/hosts` file from the previous step
