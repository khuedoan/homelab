# Roadmap

Current status: **Alpha**

## Beta requirements

Good enough for playaround with and personal use

- [x] Automated bare metal provisioning
  - [x] Controller set up (Docker)
  - [x] OS installation (PXE boot)
- [x] Automated cluster creation (k3s)
- [x] Automated application deployment (ArgoCD)
- [x] Everything is defined as code
- [ ] Basic services
  - [x] Gitea
  - [x] Tekton
  - [x] Syncthing
  - [ ] PeerTube,
  - [ ] Mail server
  - [ ] Mattermost
  - [ ] Matrix with bridges
  - [x] Vault
  - [ ] VPN
  - [ ] Dashboard
- [ ] Cloudflare tunnel or multi-cluster communication (via Wireguard or a service mesh like Linkerd)
- [ ] Local DNS
- [ ] Mirror all git repositories from GitHub automatically
- [ ] Monitoring and alerting
- [ ] Local container registry
- [ ] SSO
- [ ] Backup solution (3 copies, 2 seperate devices, 1 offsite)
- [ ] 70% availability (might break in the weekend due to new experimentation)
- [x] Only use open-source technologies

## Stable requirements

Can be used in "production" (for family or even small scale bussinesses)

- [x] A single command to deploy everything
- [x] Fast deployment time (from empty hard drive to running services under 1 hour)
- [ ] Fully _automatic_, not just _automated_
  - [ ] Bare-metal OS patching
  - [ ] Backups
  - [ ] Secrets management and rotation
  - [ ] Self healing
  - [ ] Autoscale to save electricity (optional)
- [ ] 99,9% availability (less than 9 hours of downtime per year)
- [ ] Backup encrytion
- [ ] Split DNS
- [ ] Secure by default
- [ ] Static code analysis
- [ ] Minimal dependency on external services
- [ ] Complete documentation and architecture diagram (automated update if possible)
  - [ ] Book (this book)
  - [ ] Walkthrough building tutorial and feature demo (video)

## Unplanned

Nice to have

- [ ] Addition services (TBD)
- [ ] Air-gap install
- [ ] Automated testing
- [ ] Security audit
- [ ] Migrate to RKE2 (new Terraform provider for RKE2 is not release yet)
- [ ] Serverless (OpenFaaS/Kubeless/Fission/Supabase...)
- [ ] Cluster API (https://github.com/khuedoan/homelab/pull/2)
