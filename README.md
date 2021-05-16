# Homelab

> ⚠️ WORK IN PROGRESS

## Hardware

![Hardware](https://user-images.githubusercontent.com/27996771/98970963-25137200-2543-11eb-8f2d-f9a2d45756ef.JPG)

- 4 nodes of NEC SFF PC (Japanese version of the ThinkCentre M700)
  - CPU: Intel Core i5-6600T
  - RAM: 16GB
  - SSD: 128GB
- TP-Link TL-SG108 switch

## Technology stack

<table>
  <tr>
    <td align="center"><a><img src="https://simpleicons.org/icons/ansible.svg" width="100px;"/><br/>Ansible</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/centos.svg" width="100px;"/><br/>CentOS</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/cloudflare.svg" width="100px;"/><br/>Cloudflare</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/docker.svg" width="100px;"/><br/>Docker</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/gitea.svg" width="100px;"/><br/>Gitea</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/helm.svg" width="100px;"/><br/>Helm</td>
  </tr>
  <tr>
    <td align="center"><a><img src="https://simpleicons.org/icons/kubernetes.svg" width="100px;"/><br/>Kubernetes</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/linuxcontainers.svg" width="100px;"/><br/>LXD</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/prometheus.svg" width="100px;"/><br/>Prometheus</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/rancher.svg" width="100px;"/><br/>Rancher</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/terraform.svg" width="100px;"/><br/>Terraform</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/ubuntu.svg" width="100px;"/><br/>Ubuntu</td>
  </tr>
  <tr>
    <td align="center"><a><img src="https://simpleicons.org/icons/vault.svg" width="100px;"/><br/>Vault</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/wireguard.svg" width="100px;"/><br/>Wireguard</td>
  </tr>
</table>

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
- Docker with `host` networking driver (which means [only Docker on Linux hosts](https://docs.docker.com/network/host/), you can use a Linux virtual machine with bridged networking if you're on macOS or Windows)

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
make tools
```

Then build the homelab:

```sh
make
```

## Acknowledgments

- [Fix `nf_conntrack` hash size fix for Kubernetes on LXD](https://github.com/corneliusweig/kubernetes-lxd/issues/10#issuecomment-615950053)
- [Humble project](https://github.com/locmai/humble)
- [Kubernetes on LXD issue with BTRFS](https://medium.com/@ernstae/kubenetes-on-lxd-with-rancher-2-0-part-one-and-a-half-94e6e03f4f2e)
- [LXD container profile for Kubernetes](https://github.com/justmeandopensource/kubernetes/blob/master/lxd-provisioning/k8s-profile-config)
- [Make LXD containers get IP addresses from LAN](https://blog.simos.info/how-to-make-your-lxd-container-get-ip-addresses-from-your-lan/)
- [Some device mount for Kubernetes on LXD](https://sleeplessbeastie.eu/2020/10/07/how-to-install-kubernetes-on-lxd/)
