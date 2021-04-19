# Homelab

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
- Docker with `host` networking driver (which means [only Docker on Linux hosts](https://docs.docker.com/network/host/))

For bare metal nodes:

- PXE IPv4 enabled
- Wake-on-LAN enabled
- Secure boot disabled (optional, depending on the OS)

### Configurations

- [Bare metal nodes settings](./metal/hosts.ini) (IP, MAC...)
- [OS settings](./metal/group_vars/all.yml) (PXE, network...)

### Building

Open the tools container:

```sh
make
```

Then build each layer:

```sh
make metal
make infra
make apps
```

## Release notes

<details>

None

</details>

## Roadmap

<details open>

- [ ] `0.0.1-alpha`:
  - [x] Bare metal provisioning with PXE
  - [x] LXD cluster
  - [x] Terraform state backend (etcd)
  - [x] RKE cluster
  - [ ] Core services (Vault, Gitea, DroneCI, ArgoCD,...)
  - [x] Public services to the internet (via port forwarding or Cloudflare Tunnel)
- [ ] `0.0.2-alpha`:
  - [ ] VPN (Wireguard)
  - [ ] Access the lab from the internet via VPN
  - [ ] Container registry (just pull through cache for faster cluster build time)
- [ ] `0.1.0-beta`:
  - [ ] Easy initial controller setup (with only Docker or Vagrant)
  - [ ] Fast metal image preparation
  - [ ] Mount metal image without `sudo` (7zip?)
  - [ ] Automated metal secrets generation and management
  - [ ] Automated `./infra` authentication from `./metal` (Terraform backend and provider)
  - [ ] Metal node automatic patching
  - [ ] Seperate network
  - [ ] Local DNS (PiHole?)
  - [ ] Jump box (or HashiCorp Boundary?)
  - [ ] Habor (private container registry for new applications)
  - [ ] Self managed infrastucture
  - [ ] Mirror all git repositories from GitHub automatically (with git hook for faster sync?)
  - [ ] Monitoring and alerting
  - [ ] Addition services (NextCloud, PeerTube, mailcow, Mattermost/Rocket Chat,...)
  - [ ] Dashboard for services
  - [ ] SSO
  - [ ] Backup solution (3 copies, 2 seperate devices, 1 offsite)
  - [ ] Automatic release
- [ ] `1.0.0`:
  - [ ] 100% automated (including backups and secrets management, double check with a full rebuild)
  - [ ] Cross platform inital controller support (Linux, macOS, Windows)
  - [ ] HA for everything
  - [ ] Backup encrytion
  - [ ] Secure by default
  - [ ] DRY (or rule of three)
  - [ ] Complete documentation and architecture diagram (automated update if possible)
  - [ ] Walkthrough building tutorial and feature demo
- [ ] `1.0.1`:
  - [ ] Bug fixes (TBD)
- [ ] `1.1.0`:
  - [ ] Addition services (TBD)
- [ ] Backlog:
  - [ ] Automated testing
  - [ ] Security review/audit
  - [ ] Migrate to RKE2 (new Terraform provider for RKE2 is not release yet)

</details>

You can also checkout the [TODO list](./docs/todo.md).

## Acknowledgments

- [Fix `nf_conntrack` hash size fix for Kubernetes on LXD](https://github.com/corneliusweig/kubernetes-lxd/issues/10#issuecomment-615950053)
- [Humble project](https://github.com/locmai/humble)
- [Kubernetes on LXD issue with BTRFS](https://medium.com/@ernstae/kubenetes-on-lxd-with-rancher-2-0-part-one-and-a-half-94e6e03f4f2e)
- [LXD container profile for Kubernetes](https://github.com/justmeandopensource/kubernetes/blob/master/lxd-provisioning/k8s-profile-config)
- [Make LXD containers get IP addresses from LAN](https://blog.simos.info/how-to-make-your-lxd-container-get-ip-addresses-from-your-lan/)
- [Minio Ansible role](https://github.com/atosatto/ansible-minio)
- [Some device mount for Kubernetes on LXD](https://github.com/atosatto/ansible-minio)
