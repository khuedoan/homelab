# Homelab

> ⚠️ WORK IN PROGRESS

## Hardware

![Hardware](https://user-images.githubusercontent.com/27996771/98970963-25137200-2543-11eb-8f2d-f9a2d45756ef.JPG)

- 4 nodes of NEC SFF `PC-MK26ECZDR` (Japanese version of the ThinkCentre M700):
  - CPU: `Intel Core i5-6600T @ 2.70GHz`
  - RAM: `16GB`
  - SSD: `128GB`
- TP-Link `TL-SG108` switch:
  - Ports: `8`
  - Speed: `1000Mbps`

## Architecture

A single `make` command will automatically:

- Build the `./metal` layer:
  - Create an ephemeral, stateless PXE server
  - Install Linux on all servers in parallel
- Build the `./infra` layer:
  - Create a Kubernetes [cluster](./infra/cluster.tf) using RKE
  - Install some [Helm chart for bootstrap](./infra/bootstrap.tf)
- Build the `./apps` layer:
  - Kustomize creates Argo [applications](./apps/resources)
  - ArgoCD install those applications

Visit the README file for each layer to learn more.

| Layer                  | Description                                             | Provisioner             |
|------------------------|---------------------------------------------------------|-------------------------|
| [metal](./metal)       | Bare metal OS installation, Terraform state backend,... | Ansible, PXE server     |
| [infra](./infra)       | Kubernetes cluster                                      | Terraform, Helm         |
| [apps](./apps)         | Gitea, Vault and more in the future                     | Kustomize, ArgoCD, Helm |

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

## Technology stack

<table>
  <tr>
    <td align="center"><a><img src="https://simpleicons.org/icons/ansible.svg" width="50px;"/><br/>Ansible</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/cloudflare.svg" width="50px;"/><br/>Cloudflare</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/docker.svg" width="50px;"/><br/>Docker</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/fedora.svg" width="50px;"/><br/>Fedora</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/gitea.svg" width="50px;"/><br/>Gitea</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/helm.svg" width="50px;"/><br/>Helm</td>
  </tr>
  <tr>
    <td align="center"><a><img src="https://simpleicons.org/icons/kubernetes.svg" width="50px;"/><br/>Kubernetes</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/prometheus.svg" width="50px;"/><br/>Prometheus</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/rancher.svg" width="50px;"/><br/>Rancher</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/terraform.svg" width="50px;"/><br/>Terraform</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/vault.svg" width="50px;"/><br/>Vault</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/wireguard.svg" width="50px;"/><br/>Wireguard</td>
  </tr>
  <tr>
  </tr>
</table>

## Acknowledgements

- ArgoCD usage in [my coworker's homelab](https://github.com/locmai/humble)
- [README template](https://github.com/othneildrew/Best-README-Template)
- [Run the same Cloudflare Tunnel across many `cloudflared` processes](https://developers.cloudflare.com/cloudflare-one/tutorials/many-cfd-one-tunnel)
