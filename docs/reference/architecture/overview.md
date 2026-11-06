# Overview

## Components

Main components:

- `./infra`: bare metal management, Kubernetes, and platform services
- `./apps`: user facing applications

Support components:

- `./docs`: all documentation go here, this will generate a searchable web UI
- `./tests`: automated tests for the entire system

## Provisioning flow

After editing the configuration files, run a single `just` command to:

- (1) Apply the `metal` and `cluster` units:
    - Create an ephemeral, stateless PXE server
    - Install NixOS on all servers in parallel
    - Build a Kubernetes cluster (based on k3s)
    - Export and verify its kubeconfig
- (2) Bootstrap GitOps with ArgoCD
- (3) Deploy platform components
- (4) Push the GitOps repo to Forgejo
- (5) Deploy applications in `./apps` with ArgoCD

```mermaid
flowchart BT
  cluster@{ shape: processes, label: "Servers<br/>NixOS + Kubernetes" } -->|Bootstrap| argocd[Argo CD]
  argocd -->|Deploy| platform@{ shape: processes, label: "Platform services<br/>Networking, storage, identity and secrets" }
  argocd -->|Deploy| apps@{ shape: processes, label: "Applications" }
  platform -->|Support| apps

  repository@{ shape: docs, label: "GitOps repository<br/>Application configuration" } --> argocd
  cloudflare[DNS and Tunnel] -.->|DNS and connectivity| platform
  backup[(Backup bucket)] -.->|Backup and restore| platform
```
