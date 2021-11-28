# Try on a VM

## Prerequisites

Install the following packages:

- `virtualbox`
- `vagrant`

## Build

Follow the [configuration guide](./deployment/configuration.md), then build the cluster and bootstrap it:

```sh
make -C metal dev
make bootstrap
```

Get the Ingress addresses and IPs:

```sh
export KUBECONFIG=$PWD/metal/kubeconfig.yaml
kubectl get ingress -A
```

Then update your `/etc/hosts` or your router DNS config based on the output:

```
192.168.1.150 argocd.mydomain.com
192.168.1.150 git.mydomain.com
192.168.1.150 jellyfin.mydomain.com
# etc.
```

## Clean up

Shut down the VM:

```sh
cd metal
vagrant destroy -f
```

Then clean up your router config or `/etc/hosts` file from the previous step
