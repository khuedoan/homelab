# Homelab

> ⚠️ WORK IN PROGRESS

## Hardware

![Hardware](https://user-images.githubusercontent.com/27996771/98970963-25137200-2543-11eb-8f2d-f9a2d45756ef.JPG)

- 4 nodes of NEC SFF PC `PC-MK26ECZDR` (Japanese version of the ThinkCentre M700)
  - CPU: Intel Core i5-6600T
  - RAM: 16GB
  - SSD: 128GB
- TP-Link TL-SG108 switch

## Technology stack

<table>
  <tr>
    <td align="center"><a><img src="https://simpleicons.org/icons/ansible.svg" width="100px;"/><br/>Ansible</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/cloudflare.svg" width="100px;"/><br/>Cloudflare</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/docker.svg" width="100px;"/><br/>Docker</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/fedora.svg" width="100px;"/><br/>Fedora</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/gitea.svg" width="100px;"/><br/>Gitea</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/helm.svg" width="100px;"/><br/>Helm</td>
  </tr>
  <tr>
    <td align="center"><a><img src="https://simpleicons.org/icons/kubernetes.svg" width="100px;"/><br/>Kubernetes</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/prometheus.svg" width="100px;"/><br/>Prometheus</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/rancher.svg" width="100px;"/><br/>Rancher</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/terraform.svg" width="100px;"/><br/>Terraform</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/vault.svg" width="100px;"/><br/>Vault</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/wireguard.svg" width="100px;"/><br/>Wireguard</td>
  </tr>
  <tr>
  </tr>
</table>

## Architecture

### Quick explanation

- Enter the tools container, which contains all the neccessary tools (see building instruction bellow)
- Run `make`
  - Ansible will render the [configuration file for each bare metal machine (like IP, hostname...) and the PXE server from templates](./metal/roles/pxe-boot/templates)
  - The tools container will create sibling containers to build a PXE server (includes DHCP, TFTP and HTTP server)
  - Ansible will [wake the machines up](./metal/roles/pxe-boot/tasks/wake.yml) using Wake on LAN
  - The machine start the boot process:
    - BIOS boot in network mode and look for DHCP server
    - DHCP server point it to the TFTP server to get boot files and boot config
    - The boot config contains parameter to get [automated OS installation config file](./metal/roles/pxe-boot/templates/http/kickstart/fedora.ks.j2)
    - The OS get installed and the machine reboots to the new operating system
  - Terraform will create a Kubernetes [cluster](./infra/main.tf)
  - ArgoCD will install the [applications](./apps/resources)

### Layers

| Layer | Name                   | Description                                             | Provisioner         |
|-------|------------------------|---------------------------------------------------------|---------------------|
| 0     | [metal](./metal)       | Bare metal OS installation, Terraform state backend,... | Ansible, PXE server |
| 1     | [infra](./infra)       | Kubernetes clusters                                     | Terraform, Helm     |
| 2     | [apps](./apps)         | Gitea, Vault and more in the future                     | Argo                |

## Get Started

### Prerequisite

For the controller (to run Ansible, stateless PXE server, Terraform...):

- SSH keys in `~/.ssh/{id_ed25519,id_ed25519.pub}` (you can generate it with `ssh-keygen -t ed25519`)
- Docker with `host` networking driver (which means [only Docker on Linux hosts](https://docs.docker.com/network/host/), you can use a Linux virtual machine with bridged networking if you're on macOS or Windows)

For bare metal nodes:

- PXE IPv4 enabled
- Wake-on-LAN enabled and boot to network mode by default if turned on via Wake-on-LAN
- Secure boot disabled (optional, depending on the OS)
- Note their MAC addresses

### Configurations

Change these configuration files to match your hardware and network setup:

- [Bare metal nodes settings](./metal/hosts.yml) (IP, MAC...)
- [OS settings](./metal/group_vars/all.yml) (PXE, network...)

### Build

Open the tools container:

```sh
make tools
```

Then build the homelab:

```sh
make
```

## Roadmap

See [roadmap](./docs/roadmap.md) and [open issues](https://github.com/khuedoan/homelab/issues) for a list of proposed features and known issues.

## Contributing

Any contributions you make are greatly appreciated (feature, bug fixes, documentation, grammar or typo fix...).

## License

Distributed under the GPLv3 License. See `LICENSE` for more information.

## Acknowledgements

- ArgoCD usage in [my coworker's homelab](https://github.com/locmai/humble)
- [README template](https://github.com/othneildrew/Best-README-Template)
