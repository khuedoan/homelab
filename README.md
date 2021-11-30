# Homelab

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

<table>
  <tr>
    <td align="center"><a><img src="https://simpleicons.org/icons/ansible.svg" width="50px;"/><br/>Ansible</td>
    <td align="center"><a><img src="https://www.docker.com/sites/default/files/d8/2019-07/Moby-logo.png" width="50px;"/><br/>Docker</td>
    <td align="center"><a><img src="https://avatars.githubusercontent.com/u/75713131?s=200&v=4" width="50px;"/><br/>Rocky Linux</td>
    <td align="center"><a><img src="https://cncf-branding.netlify.app/img/projects/k3s/icon/color/k3s-icon-color.svg" width="50px;"/><br/>K3s</td>
    <td align="center"><a><img src="https://cncf-branding.netlify.app/img/projects/kubernetes/icon/color/kubernetes-icon-color.svg" width="50px;"/><br/>Kubernetes</td>
  </tr>
  <tr>
    <td align="center"><a><img src="https://cncf-branding.netlify.app/img/projects/argo/icon/color/argo-icon-color.svg" width="50px;"/><br/>ArgoCD</td>
    <td align="center"><a><img src="https://cncf-branding.netlify.app/img/projects/helm/icon/color/helm-icon-color.svg" width="50px;"/><br/>Helm</td>
    <td align="center"><a><img src="https://cncf-branding.netlify.app/img/projects/longhorn/icon/color/longhorn-icon-color.svg" width="50px;"/><br/>Longhorn</td>
    <td align="center"><a><img src="https://cncf-branding.netlify.app/img/projects/prometheus/icon/color/prometheus-icon-color.svg" width="50px;"/><br/>Prometheus</td>
    <td align="center"><a><img src="https://simpleicons.org/icons/vault.svg" width="50px;"/><br/>Vault</td>
  </tr>
  <tr>
    <td align="center"><a><img src="https://upload.wikimedia.org/wikipedia/commons/b/bb/Gitea_Logo.svg" width="50px;"/><br/>Gitea</td>
    <td align="center"><a><img src="https://avatars.githubusercontent.com/u/47602533?s=200&v=4" width="50px;"/><br/>Tekton</td>
    <td align="center"><a><img src="https://knative.dev/docs/images/logo/rgb/knative-logo-rgb.png" width="50px;"/><br/>Knative</td>
  </tr>
</table>

_This list might be outdated, please let me know if I forgot to update it._

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
- [AWS Dex setup guide on EKS](https://aws.amazon.com/blogs/containers/using-dex-dex-k8s-authenticator-to-authenticate-to-amazon-eks/)

## Stargazers over time

[![Stargazers over time](https://starchart.cc/khuedoan/homelab.svg)](https://starchart.cc/khuedoan/homelab)
