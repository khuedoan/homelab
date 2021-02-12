# Khue's Home Lab

Work in progress

## Hardware setup

![Hardware](https://user-images.githubusercontent.com/27996771/98970963-25137200-2543-11eb-8f2d-f9a2d45756ef.JPG)

- 4 nodes of NEC SFF PC (Japanese version of the Thinkcentre M700)
  - CPU: Intel Core i5-6600T
  - RAM: 16GB
  - SSD: 128GB
- Switch TP-Link TL-SG108

## Architecture

| Layer | Name                   | Description                                 | Provisioner         |
|-------|------------------------|---------------------------------------------|---------------------|
| 0     | [metal](./metal)       | bare metal PXE boot, etcd, docker, lxd,     | Ansible, PXE server |
| 1     | [infra](./infra)       | Cloud-like infrastructure, based on KVM/LXC | Terraform           |
| 2     | [platform](./platform) | kubernetes, vault, git, ci/cd...            | Helm                |
| 3     | [apps](./apps)         | Plex, PeerTube, Nextcloud...                | ArgoCD              |
