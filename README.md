# Homelab

<!-- ANCHOR: introduction -->

This is my homelab set up, it focused on automation to simplify provisioning, upgrading, and operating self-hosted services.

<!-- ANCHOR_END: introduction -->

```diff
! ⚠️ WORK IN PROGRESS
```

## Overview

### Hardware

![Hardware](https://user-images.githubusercontent.com/27996771/98970963-25137200-2543-11eb-8f2d-f9a2d45756ef.JPG)

- 4 nodes of NEC SFF `PC-MK26ECZDR` (Japanese version of the ThinkCentre M700):
  - CPU: `Intel Core i5-6600T @ 2.70GHz`
  - RAM: `16GB`
  - SSD: `128GB`
- TP-Link `TL-SG108` switch:
  - Ports: `8`
  - Speed: `1000Mbps`

### Technology stack

| Logo                                                                                                                                   | Name          | Description                                                                                   |
| :------------------------------------------------------------------------------------------------------------------------------------: | :----------   | :-------------------------------------------------------------------------------------------- |
| <img width="32" alt="Ansible" src="https://simpleicons.org/icons/ansible.svg">                                                         | Ansible       | Automate bare metal provisioning and configuration                                            |
| <img width="32" alt="ArgoCD" src="https://cncf-branding.netlify.app/img/projects/argo/icon/color/argo-icon-color.svg">                 | ArgoCD        | GitOps tool built to deploy applications to Kubernetes                                        |
| <img width="32" alt="Cloudflare" src="https://avatars.githubusercontent.com/u/314135?s=200&v=4">                                       | Cloudflare    | DNS and Tunnel                                                                                |
| <img width="32" alt="Docker" src="https://www.docker.com/sites/default/files/d8/2019-07/Moby-logo.png">                                | Docker        | Ephermeral PXE server and convenient tools container                                          |
| <img width="32" alt="Gitea" src="https://upload.wikimedia.org/wikipedia/commons/b/bb/Gitea_Logo.svg">                                  | Gitea         | Self-hosted Git service                                                                       |
| <img width="32" alt="Grafana" src="https://grafana.com/static/img/menu/grafana2.svg">                                                  | Grafana       | Operational dashboards                                                                        |
| <img width="32" alt="Helm" src="https://cncf-branding.netlify.app/img/projects/helm/icon/color/helm-icon-color.svg">                   | Helm          | The package manager for Kubernetes                                                            |
| <img width="32" alt="K3s" src="https://cncf-branding.netlify.app/img/projects/k3s/icon/color/k3s-icon-color.svg">                      | K3s           | Lightweight distribution of Kubernetes                                                        |
| <img width="32" alt="Kubernetes" src="https://cncf-branding.netlify.app/img/projects/kubernetes/icon/color/kubernetes-icon-color.svg"> | Kubernetes    | Container-orchestration system, the backbone of this project                                  |
| <img width="32" alt="Let's Encrypt" src="https://avatars.githubusercontent.com/u/9289019?s=200&v=4">                                   | Let's Encrypt | Free, automated (via [cert-manager](https://cert-manager.io)), and open certificate authority |
| <img width="32" alt="Loki" src="https://github.com/grafana/loki/blob/main/docs/sources/logo.png?raw=true">                             | Loki          | Log aggregation system                                                                        |
| <img width="32" alt="Longhorn" src="https://cncf-branding.netlify.app/img/projects/longhorn/icon/color/longhorn-icon-color.svg">       | Longhorn      | Cloud native distributed block storage for Kubernetes                                         |
| <img width="32" alt="MetalLB" src="https://metallb.universe.tf/images/logo/metallb-white.png">                                         | MetalLB       | Bare metal load-balancer for Kubernetes                                                       |
| <img width="32" alt="NGINX" src="https://avatars.githubusercontent.com/u/1412239?s=200&v=4">                                           | NGINX         | Kubernetes Ingress Controller                                                                 |
| <img width="32" alt="Prometheus" src="https://cncf-branding.netlify.app/img/projects/prometheus/icon/color/prometheus-icon-color.svg"> | Prometheus    | Systems monitoring and alerting toolkit                                                       |
| <img width="32" alt="Rocky Linux" src="https://avatars.githubusercontent.com/u/75713131?s=200&v=4">                                    | Rocky Linux   | Base OS for Kubernetes nodes                                                                  |
| <img width="32" alt="Tekton" src="https://avatars.githubusercontent.com/u/47602533?s=200&v=4">                                         | Tekton        | Cloud native solution for building CI/CD systems                                              |
| <img width="32" alt="Vault" src="https://simpleicons.org/icons/vault.svg">                                                             | Vault         | Secrets and encryption management system                                                      |

## Get Started

- [Try it out on a VM](https://homelab.khuedoan.com/try_on_a_vm) without any hardware
- [Deploy on real hardware](https://homelab.khuedoan.com/deployment) for real workload

## Roadmap

See [roadmap](https://homelab.khuedoan.com/roadmap) and [open issues](https://github.com/khuedoan/homelab/issues) for a list of proposed features and known issues.

## Contributing

Any contributions you make are greatly appreciated (feature, bug fix, documentation, grammar or typo fix...).

## License

Distributed under the GPLv3 License. See `LICENSE` for more information.

## Acknowledgements

- ArgoCD usage in [my coworker's homelab](https://github.com/locmai/humble)
- [README template](https://github.com/othneildrew/Best-README-Template)
- [Run the same Cloudflare Tunnel across many `cloudflared` processes](https://developers.cloudflare.com/cloudflare-one/tutorials/many-cfd-one-tunnel)
- [MAC address environment variable in GRUB config](https://askubuntu.com/questions/1272400/how-do-i-automate-network-installation-of-many-ubuntu-18-04-systems-with-efi-and)
- [Official k3s systemd service file](https://github.com/k3s-io/k3s/blob/master/k3s.service)

## Stargazers over time

[![Stargazers over time](https://starchart.cc/khuedoan/homelab.svg)](https://starchart.cc/khuedoan/homelab)
