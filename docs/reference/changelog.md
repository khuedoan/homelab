# Changelog

## v0.1.0

Finally, a beta tag!

I haven't cut a tag in years. Compared to last time,
this is basically a rewrite and will require a
complete cluster rebuild.

Notable changes:

- Replace Fedora with NixOS. We're now all-in on Nix for the entire stack.
- Replace Ansible, Docker, Kickstart with
  [Nixie](https://github.com/khuedoan/nixie), a
  purpose-built tool for provisioning NixOS on bare
  metal
- Replace Make with Just
- Replace Terraform and glue scripts with OpenTofu and
  Terragrunt
- Bare metal and cloud infrastructure are now unified
  under the same `infra/`, with similar
  tooling (the trade-off is that external resources
  are no longer optional).
- Use a trunk-based structure for multiple
  environments (production and staging)
- Replace Terraform Cloud with Cloudflare R2 for state
  storage (you can use any S3-compatible storage)
- Use an external password manager to bootstrap
  provider secrets (`pass` by default)
- Remove the local dev cluster
- Nix is now required on the initial controller for
  better integration
- Expand monitoring and alerting, add ntfy for
  notifications
- Coordinate NixOS reboots with Kured
- Add Istio ambient mesh and enforce mTLS for
  applications
- Replace NGINX Ingress with Istio and Gateway API
- Add Kiali for network visualization
- Replace MetalLB with kube-vip
- Add Tailscale and WireGuard Portal, with SSO and
  self-service device setup in the portal
- Replace Dex with Kanidm for SSO
- Automate provisioning for Kanidm users, groups, and
  OAuth clients
- Replace Longhorn with Rook Ceph for distributed
  storage
- Add Velero for scheduled offsite backups to
  S3-compatible storage (Backblaze B2 by default)
- Add cluster recovery mode for use during restores
- Replace Vault with OpenBao
- Add namespace-scoped access to secrets
- Replace Gitea with Forgejo
- Replace Tekton with Forgejo Actions
- Add shared `linux` Forgejo runners with Nix support
- Replace Trow with Zot
- Move Renovate from a standalone Kubernetes CronJob
  to Forgejo Actions
- Federate Forgejo Actions with Zot so workflows can
  publish to their scoped paths
- Enforce the `app-template` chart in `apps/`, with
  platform guardrails for applications
- Add Actual Budget, Immich, PairDrop, Excalidraw,
  Hister, OpenSpeedTest
- Replace Hajimari with Glance. I don't really need
  automated ingress discovery
- Replace Seafile with OpenCloud
- Remove Paperless, Jellyfin, and other media tools
  (at least for now)
- A more complete test suite for networking, storage,
  backup and restore, secret access, GitOps, and CI
  workflows that build, push, and run container images
- Add benchmarks for storage and K3s security
- More docs!

Experiments I decided to remove:

- Homepage: I prefer Glance's minimalism
- Ollama: my laptops are way more powerful than my
  homelab nodes and can run local models better
- Woodpecker: no workflow OIDC
- Harbor and Docker Registry: Zot is lightweight and
  has the features that I need (OIDC, UI, pull-through
  cache, etc.)
- Cilium + Hubble: Istio + Kiali is more portable, and
  I already run Istio in production
- ZeroTier: WireGuard Portal + Tailscale strike a
  better balance by offering good UX without lock-in
- Shared Kubernetes "global" Secret: I still need an
  API for dynamic secrets and access control. OpenBao
  is basically an open-source drop-in replacement for
  Vault
- VolSync: Velero makes it easier to build a "back
  everything up" experience

## v0.0.8

Notable changes:

- **build:** run post install scripts by default
- **build:** set `KUBECONFIG` from global Makefile
- **feat(external-dns)!:** add cluster name as owner ID
- **feat(tools):** install `yamllint`, `ansible-lint` and `k9s`
- **feat(tools):** set `KUBECONFIG` by default
- **feat:** add pre-commit hooks
- **feat:** add script to setup Gitea tokens and OAuth apps
- **perf(argocd):** turning on selective sync
- **refactor(docs):** migrate to [mkdocs](https://squidfunk.github.io/mkdocs-material)
- **refactor(metal):** migrate to Fedora 36 for newer packages
- **refactor(pxe)!:** combine dhcpd and tftpd to dnsmasq
- Many bug fixes

Please see git log for full change log.

## 0.0.7-alpha

- Replace standard Vault with Vault Operator
- Automatically initialize and unseal Vault
- Declarative secret generation and management
- Declarative Gitea configuration with YAML
- Automatic OS rolling upgrade
- Automatic Kubernetes rolling upgrade
- Automatic application updates using Renovate (still require manual token generation)
- Add script to wait for essential services after deployment
- Add icons and bookmarks to the home page
- Deploy Matrix chat
- Replace Authentik with Dex for SSO (still require manual token generation)
- Switch to Mermaid for diagrams in documentation
- Replace Vagrant with k3d for development environment
- Use nip.io domain for development environment
- Remove Backblaze (S3 Glacier and/or Minio will be added in future version)
- Enable monitor for the majority of applications
- Many code refactorings and bug fixes

## 0.0.6-alpha

- Upgrade to Kubernetes 1.23
- Support external resources:
    - Cloudflare DNS and Tunnel
    - Backblaze for backup
    - Auto inject secrets to required namespaces
- Replace self-signed certificates with Let's Encrypt production (with API token injected from the `external` layer)
- Add DNS records automatically using external-dns
- Easy Cloudflare Tunnel configuration with annotations
- Offsite backup to Backblaze B2 bucket using k8up-operator
- Add private container registry
- Remove Knative to save resources (temporarily)
- Enable encryption at rest for Kubernetes Secrets
- Add more Tekton tasks and pipelines
- Initialize GitOps repository on Gitea automatically after install
- Generate MetalLB address pool automatically (default to the last `/27` subnet)
- Some bug fixes

## 0.0.5-alpha

- Add convenience scripts
- Add Loki for logging
- Add custom health check for Application and ApplicationSet
- Use Vault with dev mode on (temporarily until we hit beta)
- Replace Authelia with Authentik
- Upgrade to Kubernetes 1.22
- Upgrade most services to the latest version
- Set ingress class and storage class explicitly
- Initial Linkerd and Knative setup (not working yet)
- Set up Hajimari for home page with automatic ingress discovery
- Add dev VM for local development or evaluation
- Optimize bare metal provisioning performance
- Replace Syncthing with Seafile (may use both in the feature)
- Enable Gitea SSH cloning via Ingress
- Various code clean up
- Add more documents

## 0.0.4-alpha

- Switch to Rocky Linux
- Some optimization for bare metal provisioning
- Switch to k3s and combine Kubernetes cluster config in `./infra` layer to `./metal` layer (because k3s is also configured using Ansible)
- Split boostrap Helm charts in `./infra` layer to `./bootstrap` layer (with new ArgoCD pattern) and `./system` layer
- Add `./platform` layer for some applications like Gitea, Tekton...
- User only need to provision `./metal` and `bootstrap` layer, the `./bootstrap` layer will deploy the remaining layers
- Provisioning time from empty disk to running services is significantly reduced (thanks to k3s and new bootstrap pattern)
- Use [mdBook](https://rust-lang.github.io/mdBook/) for documents
- Replace Drone CI with Tekton
- Enable TLS on all Ingresses (using [cert-manager](https://cert-manager.io))
- Add some new applications

## 0.0.3-alpha

- Generate Terraform backend config automatically
- Switch to CoreOS
- Better PXE boot setup
- Diagrams as code

## 0.0.2-alpha

- Ensure idempotency for bare metal provisioning
- Extract instead of mounting the OS ISO file
- Easy initial controller setup (with only Docker)
- Switch to Fedora
- Remove LXD
- Move etcd (Terraform state backend) back to Docker

## 0.0.1-alpha

- Bare metal provisioning with PXE
- LXD cluster
- Terraform state backend (etcd)
- RKE cluster
- Core services (Vault, Gitea, ArgoCD,...)
- Public services to the internet (via port forwarding or Cloudflare Tunnel)
