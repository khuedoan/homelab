<div align="center">

# Khue's Homelab

<!-- ANCHOR: introduction -->

[![chat](https://img.shields.io/matrix/homelab:matrix.khuedoan.com?style=flat-square&logo=matrix&logoColor=white&label=chat)](https://matrix.to/#/#homelab:matrix.khuedoan.com)
[![tag](https://img.shields.io/github/v/tag/khuedoan/homelab?style=flat-square&logo=semver&logoColor=white)](https://github.com/khuedoan/homelab/tags)
[![document](https://img.shields.io/website?label=document&logo=gitbook&logoColor=white&style=flat-square&url=https%3A%2F%2Fhomelab.khuedoan.com)](https://homelab.khuedoan.com)
[![license](https://img.shields.io/github/license/khuedoan/homelab?style=flat-square&logo=gnu&logoColor=white)](https://www.gnu.org/licenses/gpl-3.0.html)
[![stars](https://img.shields.io/github/stars/khuedoan/homelab?logo=github&logoColor=white&color=gold&style=flat-square)](https://github.com/khuedoan/homelab)

This project utilizes [Infrastructure as Code](https://en.wikipedia.org/wiki/Infrastructure_as_code) and [GitOps](https://www.weave.works/technologies/gitops) to automate provisioning, operating, and updating self-hosted services in my homelab.
It can be used as a highly customizable framework to build your own homelab.

<!-- TODO -->
<!-- Feel free to join me on my Matrix chat server at [chat.khuedoan.com](https://chat.khuedoan.com/#/room/#homelab/general:matrix.khuedoan.com), -->
<!-- or [`#homelab:matrix.khuedoan.com`](https://matrix.to/#/#homelab:matrix.khuedoan.com) if you already have a Matrix client. -->
<!-- Please note that the chat server is self-hosted on my homelab and still at the experimental stage. -->

<!-- ANCHOR_END: introduction -->

Current status: **ALPHA**

</div>

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
- [x] Automatic rolling upgrade for OS and Kubernetes
- [x] Automatically update apps (with approval)
- [x] Modular architecture, easy to add or remove features/components
- [x] Automated certificate management
- [x] Automatically update DNS records for exposed services
- [x] Expose services to the internet securely with [Cloudflare Tunnel](https://www.cloudflare.com/products/tunnel/)
- [x] CI/CD platform
- [x] Private container registry
- [x] Distributed storage
- [x] Support multiple environments (dev, prod)
- [ ] Monitoring and alerting 🚧
- [ ] Automated offsite backups 🚧
- [ ] Single sign-on 🚧

Some demo videos and screenshots are shown here.
They can't capture all of the project's features, but they are sufficient to get a concept of it.

| [![Deployment](https://asciinema.org/a/xkBRkwC6e9RAzVuMDXH3nGHp7.svg)](https://asciinema.org/a/xkBRkwC6e9RAzVuMDXH3nGHp7) |
| :--:                                                                                                                      |
| Deploy with a single command (after updating the config files of course)                                                  |

| [![PXE boot](https://user-images.githubusercontent.com/27996771/157303477-df2e7410-8f02-4648-a86c-71e6b7e89e35.png)](https://www.youtube.com/watch?v=y-d7btNNAT8) |
| :--:                                                                                                                                                              |
| PXE boot                                                                                                                                                          |

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

<table>
  <tr>
    <th>Logo</th>
    <th>Name</th>
    <th>Description</th>
  </tr>
  <tr>
    <td><img width="32" src="https://simpleicons.org/icons/ansible.svg"></td>
    <td><a href="https://www.ansible.com">Ansible</a></td>
    <td>Automate bare metal provisioning and configuration</td>
  </tr>
  <tr>
    <td><img width="32" src="https://cncf-branding.netlify.app/img/projects/argo/icon/color/argo-icon-color.svg"></td>
    <td><a href="https://argoproj.github.io/cd">ArgoCD</a></td>
    <td>GitOps tool built to deploy applications to Kubernetes</td>
  </tr>
  <tr>
    <td><img width="32" src="https://github.com/jetstack/cert-manager/raw/master/logo/logo.png"></td>
    <td><a href="https://cert-manager.io">cert-manager</a></td>
    <td>Cloud native certificate management</td>
  </tr>
  <tr>
    <td><img width="32" src="https://avatars.githubusercontent.com/u/314135?s=200&v=4"></td>
    <td><a href="https://www.cloudflare.com">Cloudflare</a></td>
    <td>DNS and Tunnel</td>
  </tr>
  <tr>
    <td><img width="32" src="https://www.docker.com/wp-content/uploads/2022/03/Moby-logo.png"></td>
    <td><a href="https://www.docker.com">Docker</a></td>
    <td>Ephermeral PXE server and convenient tools container</td>
  </tr>
  <tr>
    <td><img width="32" src="https://github.com/kubernetes-sigs/external-dns/raw/master/img/external-dns.png"></td>
    <td><a href="https://github.com/kubernetes-sigs/external-dns">ExternalDNS</a></td>
    <td>Synchronizes exposed Kubernetes Services and Ingresses with DNS providers</td>
  </tr>
  <tr>
    <td><img width="32" src="https://upload.wikimedia.org/wikipedia/commons/b/bb/Gitea_Logo.svg"></td>
    <td><a href="https://gitea.com">Gitea</a></td>
    <td>Self-hosted Git service</td>
  </tr>
  <tr>
    <td><img width="32" src="https://grafana.com/static/img/menu/grafana2.svg"></td>
    <td><a href="https://grafana.com">Grafana</a></td>
    <td>Operational dashboards</td>
  </tr>
  <tr>
    <td><img width="32" src="https://cncf-branding.netlify.app/img/projects/helm/icon/color/helm-icon-color.svg"></td>
    <td><a href="https://helm.sh">Helm</a></td>
    <td>The package manager for Kubernetes</td>
  </tr>
  <tr>
    <td><img width="32" src="https://cncf-branding.netlify.app/img/projects/k3s/icon/color/k3s-icon-color.svg"></td>
    <td><a href="https://k3s.io">K3s</a></td>
    <td>Lightweight distribution of Kubernetes</td>
  </tr>
  <tr>
    <td><img width="32" src="https://cncf-branding.netlify.app/img/projects/kubernetes/icon/color/kubernetes-icon-color.svg"></td>
    <td><a href="https://kubernetes.io">Kubernetes</a></td>
    <td>Container-orchestration system, the backbone of this project</td>
  </tr>
  <tr>
    <td><img width="32" src="https://github.com/grafana/loki/blob/main/docs/sources/logo.png?raw=true"></td>
    <td><a href="https://grafana.com/oss/loki">Loki</a></td>
    <td>Log aggregation system</td>
  </tr>
  <tr>
    <td><img width="32" src="https://cncf-branding.netlify.app/img/projects/longhorn/icon/color/longhorn-icon-color.svg"></td>
    <td><a href="https://longhorn.io">Longhorn</a></td>
    <td>Cloud native distributed block storage for Kubernetes</td>
  </tr>
  <tr>
    <td><img width="32" src="https://avatars.githubusercontent.com/u/60239468?s=200&v=4"></td>
    <td><a href="https://metallb.org">MetalLB</a></td>
    <td>Bare metal load-balancer for Kubernetes</td>
  </tr>
  <tr>
    <td><img width="32" src="https://avatars.githubusercontent.com/u/1412239?s=200&v=4"></td>
    <td><a href="https://www.nginx.com">NGINX</a></td>
    <td>Kubernetes Ingress Controller</td>
  </tr>
  <tr>
    <td><img width="32" src="https://cncf-branding.netlify.app/img/projects/prometheus/icon/color/prometheus-icon-color.svg"></td>
    <td><a href="https://prometheus.io">Prometheus</a></td>
    <td>Systems monitoring and alerting toolkit</td>
  </tr>
  <tr>
    <td><img width="32" src="https://docs.renovatebot.com/assets/images/logo.png"></td>
    <td><a href="https://www.whitesourcesoftware.com/free-developer-tools/renovate">Renovate</a></td>
    <td>Automatically update dependencies</td>
  </tr>
  <tr>
    <td><img width="32" src="https://avatars.githubusercontent.com/u/75713131?s=200&v=4"></td>
    <td><a href="https://rockylinux.org">Rocky Linux</a></td>
    <td>Base OS for Kubernetes nodes</td>
  </tr>
  <tr>
    <td><img width="32" src="https://avatars.githubusercontent.com/u/47602533?s=200&v=4"></td>
    <td><a href="https://tekton.dev">Tekton</a></td>
    <td>Cloud native solution for building CI/CD systems</td>
  </tr>
  <tr>
    <td><img width="32" src="https://trow.io/trow.png"></td>
    <td><a href="https://trow.io">Trow</a></td>
    <td>Private container registry</td>
  </tr>
  <tr>
    <td><img width="32" src="https://simpleicons.org/icons/vault.svg"></td>
    <td><a href="https://www.vaultproject.io">Vault</a></td>
    <td>Secrets and encryption management system</td>
  </tr>
</table>

## Get Started

- [Try it out locally](https://homelab.khuedoan.com/try-locally.html) without any hardware
- [Deploy on real hardware](https://homelab.khuedoan.com/deployment) for real workload

## Roadmap

See [roadmap](https://homelab.khuedoan.com/roadmap.html) and [open issues](https://github.com/khuedoan/homelab/issues) for a list of proposed features and known issues.

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
