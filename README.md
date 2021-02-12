# Khue's Home Lab

Work in progress

## Hardware

![Hardware](https://user-images.githubusercontent.com/27996771/98970963-25137200-2543-11eb-8f2d-f9a2d45756ef.JPG)

- 4 nodes of NEC SFF PC (Japanese version of the ThinkCentre M700)
  - CPU: Intel Core i5-6600T
  - RAM: 16GB
  - SSD: 128GB
- TP-Link TL-SG108 switch

## Architecture

| Layer | Name                   | Description                                 | Provisioner         |
|-------|------------------------|---------------------------------------------|---------------------|
| 0     | [metal](./metal)       | Bare metal OS installation, Docker, tfstate | Ansible, PXE server |
| 1     | [infra](./infra)       | Kubernetes cluster                          | Terraform           |
| 2     | [platform](./platform) | Vault, Git, Argo,...                        | Helm                |
| 3     | [apps](./apps)         |                                             | ArgoCD              |
