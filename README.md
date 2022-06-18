<div align="center">

# Pando85's Homelab

<!-- ANCHOR: introduction -->

[![tag](https://img.shields.io/github/v/tag/pando85/homelab?style=flat-square&logo=semver&logoColor=white)](https://github.com/pando85/homelab/tags)
[![document](https://img.shields.io/website?label=document&logo=gitbook&logoColor=white&style=flat-square&url=https%3A%2F%2Fhomelab.pando85.com)](https://homelab.pando85.com)
[![license](https://img.shields.io/github/license/pando85/homelab?style=flat-square&logo=gnu&logoColor=white)](https://www.gnu.org/licenses/gpl-3.0.html)
[![stars](https://img.shields.io/github/stars/pando85/homelab?logo=github&logoColor=white&color=gold&style=flat-square)](https://github.com/pando85/homelab)

This project utilizes [Infrastructure as Code](https://en.wikipedia.org/wiki/Infrastructure_as_code) and [GitOps](https://www.weave.works/technologies/gitops) to automate provisioning, operating, and updating self-hosted services in my homelab.
It can be used as a highly customizable framework to build your own homelab.

<!-- ANCHOR_END: introduction -->

Current status: **ALPHA**

</div>

## Overview

This section provides a high level overview of the project.
For further information, please see the [documentation](TODO).

### Hardware

TODO
### Features

Project status: **Alpha** (see [roadmap](#roadmap) below)

- [x] Common applications: Gitea...
- [x] Automated Kubernetes installation and management
- [x] Installing and managing applications using GitOps
- [x] Automatic rolling upgrade for OS and Kubernetes
- [x] Automatically update apps (with approval)
- [x] Modular architecture, easy to add or remove features/components
- [x] Automated certificate management
- [x] CI/CD platform
- [x] Private container registry
- [x] Distributed storage
- [x] Support multiple environments (dev, master)
- [ ] Automatically update DNS records for exposed services
- [ ] Monitoring and alerting 🚧
- [ ] Automated offsite backups 🚧
- [ ] Single sign-on 🚧

## License

<details>

<summary>Distributed under the GPLv3 License.</summary>

This project is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This project is distributed in the hope that it will be useful, but **WITHOUT ANY WARRANTY**; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this project (`LICENSE.md`).
If not, see <https://www.gnu.org/licenses>.

</details>

## Acknowledgements

- [khuedoan/homelab](https://github.com/khuedoan/homelab)
- [ArgoCD usage in my coworker's homelab](https://github.com/locmai/humble)
- [README template](https://github.com/othneildrew/Best-README-Template)
- [Run the same Cloudflare Tunnel across many `cloudflared` processes](https://developers.cloudflare.com/cloudflare-one/tutorials/many-cfd-one-tunnel)
- [MAC address environment variable in GRUB config](https://askubuntu.com/questions/1272400/how-do-i-automate-network-installation-of-many-ubuntu-18-04-systems-with-efi-and)
- [Official k3s systemd service file](https://github.com/k3s-io/k3s/blob/master/k3s.service)
- [Official Cloudflare Tunnel examples](https://github.com/cloudflare/argo-tunnel-examples)
- [Initialize GitOps repository on Gitea and integrate with Tekton by RedHat](https://github.com/redhat-scholars/tekton-tutorial/tree/master/triggers)
- [SSO configuration from xUnholy/k8s-gitops](https://github.com/xUnholy/k8s-gitops)
