# Khue's Homelab

<!-- ANCHOR: introduction -->

[![chat](https://img.shields.io/matrix/homelab:matrix.khuedoan.com?style=flat-square&logo=matrix&logoColor=white&label=chat)](https://matrix.to/#/#homelab:matrix.khuedoan.com)
[![tag](https://img.shields.io/github/v/tag/khuedoan/homelab?style=flat-square&logo=semver&logoColor=white)](https://github.com/khuedoan/homelab/tags)
[![document](https://img.shields.io/website?label=document&logo=gitbook&logoColor=white&style=flat-square&url=https%3A%2F%2Fhomelab.khuedoan.com)](https://homelab.khuedoan.com)
[![license](https://img.shields.io/github/license/khuedoan/homelab?style=flat-square&logo=gnu&logoColor=white)](https://www.gnu.org/licenses/gpl-3.0.html)
[![stars](https://img.shields.io/github/stars/khuedoan/homelab?logo=github&logoColor=white&color=gold&style=flat-square)](https://github.com/khuedoan/homelab)

Current status: **ALPHA**

This project utilizes [Infrastructure as Code](https://en.wikipedia.org/wiki/Infrastructure_as_code) to automate provisioning, operating, and updating self-hosted services in my homelab.
It can be used as a highly customizable framework to build your own homelab.

<!-- TODO -->
<!-- Feel free to join me on my Matrix chat server at [chat.khuedoan.com](https://chat.khuedoan.com/#/room/#homelab/general:matrix.khuedoan.com), -->
<!-- or [`#homelab:matrix.khuedoan.com`](https://matrix.to/#/#homelab:matrix.khuedoan.com) if you already have a Matrix client. -->
<!-- Please note that the chat server is self-hosted on my homelab and still at the experimental stage. -->

<!-- ANCHOR_END: introduction -->

## Overview

This section provides a high level overview of the project.
For further information, please see the [documentation](https://homelab.khuedoan.com).

### Hardware

![Hardware](https://user-images.githubusercontent.com/27996771/98970963-25137200-2543-11eb-8f2d-f9a2d45756ef.JPG)

- 4 × NEC SFF `PC-MK26ECZDR` (Japanese version of the ThinkCentre M700):
  - CPU: `Intel Core i5-6600T @ 2.70GHz`
  - RAM: `16GB`
  - SSD: `128GB`
- TP-Link `TL-SG108` switch:
  - Ports: `8`
  - Speed: `1000Mbps`

### Features

Project status: **Alpha** (see [roadmap](#roadmap) below)

- [x] Common applications: Gitea, Seafile, Jellyfin, Paperless...
- [x] Automated bare metal provisioning with PXE boot
- [x] Automated Kubernetes installation and management
- [x] Installing and managing applications using GitOps
- [x] Modular architecture, easy to add or remove features/components
- [x] Automated certificate management
- [x] Automatically update DNS records for exposed services
- [x] Expose services to the internet securely with [Cloudflare Tunnel](https://www.cloudflare.com/products/tunnel/)
- [x] CI/CD platform
- [x] Private container registry
- [x] Distributed storage
- [ ] Monitoring and alerting 🚧
- [ ] Support multiple environments (dev, stag, prod) 🚧
- [ ] Automated offsite backups 🚧
- [ ] Single sign-on 🚧

Screenshots of some user-facing applications are shown here, I will update them before each release.
They can't capture all of the project's features, but they are sufficient to get a concept of it.

| ![](https://user-images.githubusercontent.com/27996771/149445807-0f869eb7-d8f5-4fef-ab97-ac281df91a06.png) |
| :--:                                                                                                       |
| Homepage with Ingress discovery powered by [Hajimari](https://github.com/toboshii/hajimari)                |

| ![](https://user-images.githubusercontent.com/27996771/149444871-38889c9d-862f-41ff-8c05-8ece21da3e9c.png) |
| :--:                                                                                                       |
| Git server powered by [Gitea](https://gitea.io/en-us/)                                                     |

| ![](https://user-images.githubusercontent.com/27996771/149445374-58fd0605-bb9a-46e4-81d6-5e584d2b94a9.png) |
| :--:                                                                                                       |
| Continuous integration with [Tekton](https://tekton.dev/)                                                  |

| ![](https://user-images.githubusercontent.com/27996771/149444716-fc0d7282-4cf7-4ddb-97a4-1a3fb47ff2b8.png) |
| :--:                                                                                                       |
| Continuous deployment with [ArgoCD](https://argoproj.github.io/cd/)                                        |

| ![](https://user-images.githubusercontent.com/27996771/149446631-1c5d056b-1fdc-48e6-96ba-e1abe1762be0.png) |
| :--:                                                                                                       |
| Monitoring dashboard powered by [Grafana](https://grafana.com/)                                            |

| ![](https://user-images.githubusercontent.com/27996771/149448510-7163310c-2049-4ccd-901d-f11f605bfc32.png)                                                                       |
| :--:                                                                                                                                                                             |
| [Matrix](https://matrix.org/) chat server powered by [Element](https://matrix.org/docs/projects/client/element) and [Dendrite](https://matrix.org/docs/projects/server/dendrite) |

| ![](https://user-images.githubusercontent.com/27996771/149448896-9d79947d-468c-45c6-a81d-b43654e8ab6b.png) |
| :--:                                                                                                       |
| Cluster management using [Lens](https://k8slens.dev/) (or you can just use `kubectl`)                      |

| ![](https://user-images.githubusercontent.com/27996771/149452309-de4a893b-e94c-4ba8-9119-ea87449cf77e.png) |
| :--:                                                                                                       |
| Secret management with [Vault](https://www.vaultproject.io/)                                               |

### Tech stack

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
| <img width="32" alt="MetalLB" src="https://avatars.githubusercontent.com/u/60239468?s=200&v=4">                                        | MetalLB       | Bare metal load-balancer for Kubernetes                                                       |
| <img width="32" alt="NGINX" src="https://avatars.githubusercontent.com/u/1412239?s=200&v=4">                                           | NGINX         | Kubernetes Ingress Controller                                                                 |
| <img width="32" alt="Prometheus" src="https://cncf-branding.netlify.app/img/projects/prometheus/icon/color/prometheus-icon-color.svg"> | Prometheus    | Systems monitoring and alerting toolkit                                                       |
| <img width="32" alt="Rocky Linux" src="https://avatars.githubusercontent.com/u/75713131?s=200&v=4">                                    | Rocky Linux   | Base OS for Kubernetes nodes                                                                  |
| <img width="32" alt="Tekton" src="https://avatars.githubusercontent.com/u/47602533?s=200&v=4">                                         | Tekton        | Cloud native solution for building CI/CD systems                                              |
| <img width="32" alt="Trow" src="https://trow.io/trow.png">                                                                             | Trow          | Private container registry                                                                    |
| <img width="32" alt="Vault" src="https://simpleicons.org/icons/vault.svg">                                                             | Vault         | Secrets and encryption management system                                                      |

## Get Started

- [Try it out on a VM](https://homelab.khuedoan.com/try_on_a_vm) without any hardware
- [Deploy on real hardware](https://homelab.khuedoan.com/deployment) for real workload

## Roadmap

See [roadmap](https://homelab.khuedoan.com/roadmap) and [open issues](https://github.com/khuedoan/homelab/issues) for a list of proposed features and known issues.

## Contributing

Any contributions you make, either big or small, are greatly appreciated.

## License

> Copyright (c) 2020, 2021, 2022 Khue Doan

<details>

<summary>Distributed under the GPLv3 License.</summary>

This project is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This project is distributed in the hope that it will be useful, but **WITHOUT ANY WARRANTY**; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this project (`LICENSE.md`).
If not, see <https://www.gnu.org/licenses>.

</details>

## Acknowledgements

- [ArgoCD usage in my coworker's homelab](https://github.com/locmai/humble)
- [README template](https://github.com/othneildrew/Best-README-Template)
- [Run the same Cloudflare Tunnel across many `cloudflared` processes](https://developers.cloudflare.com/cloudflare-one/tutorials/many-cfd-one-tunnel)
- [MAC address environment variable in GRUB config](https://askubuntu.com/questions/1272400/how-do-i-automate-network-installation-of-many-ubuntu-18-04-systems-with-efi-and)
- [Official k3s systemd service file](https://github.com/k3s-io/k3s/blob/master/k3s.service)
- [Official Cloudflare Tunnel examples](https://github.com/cloudflare/argo-tunnel-examples)
- [Initialize GitOps repository on Gitea and integrate with Tekton by RedHat](https://github.com/redhat-scholars/tekton-tutorial/tree/master/triggers)

## Stargazers over time

[![Stargazers over time](https://starchart.cc/khuedoan/homelab.svg)](https://starchart.cc/khuedoan/homelab)
