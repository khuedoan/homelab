# Pando85's Homelab

**[Features](#features) • [Get Started](#get-started) • [Documentation](https://homelab.khuedoan.com)**

[![tag](https://img.shields.io/github/v/tag/pando85/homelab?style=flat-square&logo=semver&logoColor=white)](https://github.com/pando85/homelab/tags)
[![document](https://img.shields.io/website?label=document&logo=gitbook&logoColor=white&style=flat-square&url=https%3A%2F%2Fhomelab.pando85.com)](https://homelab.pando85.com)
[![license](https://img.shields.io/github/license/pando85/homelab?style=flat-square&logo=gnu&logoColor=white)](https://www.gnu.org/licenses/gpl-3.0.html)
[![stars](https://img.shields.io/github/stars/pando85/homelab?logo=github&logoColor=white&color=gold&style=flat-square)](https://github.com/pando85/homelab)

This project utilizes [Infrastructure as Code](https://en.wikipedia.org/wiki/Infrastructure_as_code) and [GitOps](https://www.weave.works/technologies/gitops) to automate provisioning, operating, and updating self-hosted services in my homelab.
It can be used as a highly customizable framework to build your own homelab.

> **What is a homelab?**
>
> Homelab is a laboratory at home where you can self-host, experiment with new technologies, practice for certifications, and so on.
> For more information about homelab in general, see the [r/homelab introduction](https://www.reddit.com/r/homelab/wiki/introduction).

## Overview

<<<<<<< HEAD
This section provides a high level overview of the project.
For further information, please see the [documentation](TODO).

### Hardware

TODO
### Features

Project status: **Alpha** (see [roadmap](#roadmap) below)

- [x] Common applications: Gitea...
=======
Project status: **ALPHA**

This project is still in the experimental stage, and I don't use anything critical on it.
Expect breaking changes that may require a complete redeployment.
A proper upgrade path is planned for the stable release.
More information can be found in [the roadmap](#roadmap) below.

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

- [x] Common applications: Gitea, Seafile, Jellyfin, Paperless...
- [x] Automated bare metal provisioning with PXE boot
>>>>>>> c14a611bfad59061f8546409ac5f940186bf24e9
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
- [ArgoCD usage and monitoring configuration in locmai/humble](https://github.com/locmai/humble)
>>>>>>> c14a611bfad59061f8546409ac5f940186bf24e9
- [README template](https://github.com/othneildrew/Best-README-Template)
- [Run the same Cloudflare Tunnel across many `cloudflared` processes](https://developers.cloudflare.com/cloudflare-one/tutorials/many-cfd-one-tunnel)
- [MAC address environment variable in GRUB config](https://askubuntu.com/questions/1272400/how-do-i-automate-network-installation-of-many-ubuntu-18-04-systems-with-efi-and)
- [Official k3s systemd service file](https://github.com/k3s-io/k3s/blob/master/k3s.service)
- [Official Cloudflare Tunnel examples](https://github.com/cloudflare/argo-tunnel-examples)
- [Initialize GitOps repository on Gitea and integrate with Tekton by RedHat](https://github.com/redhat-scholars/tekton-tutorial/tree/master/triggers)
- [SSO configuration from xUnholy/k8s-gitops](https://github.com/xUnholy/k8s-gitops)
- [Pre-commit config from k8s-at-home/flux-cluster-template](https://github.com/k8s-at-home/flux-cluster-template)

Here is a list of the contributors who have helped to improve this project.
Big shout-out to them!

- ![](https://github.com/locmai.png?size=24) [@locmai](https://github.com/locmai)
- ![](https://github.com/MatthewJohn.png?size=24) [@MatthewJohn](https://github.com/MatthewJohn)
- ![](https://github.com/karpfediem.png?size=24) [@karpfediem](https://github.com/karpfediem)
- ![](https://github.com/linhng98.png?size=24) [@linhng98](https://github.com/linhng98)
- ![](https://github.com/BlueHatbRit.png?size=24) [@BlueHatbRit](https://github.com/BlueHatbRit)
- ![](https://github.com/dotdiego.png?size=24) [@dotdiego](https://github.com/dotdiego)
- ![](https://github.com/Crimrose.png?size=24) [@Crimrose](https://github.com/Crimrose)
- ![](https://github.com/eventi.png?size=24) [@eventi](https://github.com/eventi)
- ![](https://github.com/Bourne-ID.png?size=24) [@Bourne-ID](https://github.com/Bourne-ID)

If you feel you're missing from this list, feel free to add yourself in a PR.

