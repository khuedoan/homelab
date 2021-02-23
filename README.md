# Khue's homelab

```diff
! WORK IN PROGRESS
```

## Hardware

![Hardware](https://user-images.githubusercontent.com/27996771/98970963-25137200-2543-11eb-8f2d-f9a2d45756ef.JPG)

- 4 nodes of NEC SFF PC (Japanese version of the ThinkCentre M700)
  - CPU: Intel Core i5-6600T
  - RAM: 16GB
  - SSD: 128GB
- TP-Link TL-SG108 switch

## Architecture

| Layer | Name                   | Description                                                  | Provisioner         |
|-------|------------------------|--------------------------------------------------------------|---------------------|
| 0     | [metal](./metal)       | Bare metal OS installation, LXD, Terraform state backend,... | Ansible, PXE server |
| 1     | [infra](./infra)       | Kubernetes clusters, shared apps (Git, Vault, Argo...)       | Terraform, Helm     |
| 2     | [apps](./apps)         |                                                              | Argo                |

## Usage

### Prerequisite

For the controller (to run Ansible, stateless PXE server, Terraform...):

- SSH keys in `~/.ssh/{id_rsa,id_rsa.pub}`
- `make`
- `python3`
- Docker with `host` networking driver (which means [only Docker on Linux hosts](https://docs.docker.com/network/host/))
- `terraform` (0.14.x)
- `vagrant` (optional, to create a local [test environment](./test))

For bare metal nodes:

- PXE IPv4 enabled
- Wake-on-LAN enabled
- Secure boot disabled (optional, depending on the OS)

### Configurations

- [Bare metal nodes settings](./metal/hosts.ini) (IP, MAC...)
- [OS settings](./metal/group_vars/all.yml) (PXE, network...)

### Building

Simply run:

```sh
make
```

Or we can build each layer individually:

```sh
make infra
# or
cd infra
make
```

## Acknowledgements

- [Fix `nf_conntrack` hash size fix for Kubernetes on LXD](https://github.com/corneliusweig/kubernetes-lxd/issues/10#issuecomment-615950053)
- [Humble project](https://github.com/locmai/humble)
- [Kubernetes on LXD issue with BTRFS](https://medium.com/@ernstae/kubenetes-on-lxd-with-rancher-2-0-part-one-and-a-half-94e6e03f4f2e)
- [LXD container profile for Kubernetes](https://github.com/justmeandopensource/kubernetes/blob/master/lxd-provisioning/k8s-profile-config)
- [Minio Ansible role](https://github.com/atosatto/ansible-minio)
- [Some device mount for Kubernetes on LXD](https://github.com/atosatto/ansible-minio)
