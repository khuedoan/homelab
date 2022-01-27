# Roadmap

> Current status: **Alpha**

## Alpha requirements

- [x] Literally anything that works

## Beta requirements

Good enough for playaround with and personal use

- [x] Automated bare metal provisioning
  - [x] Controller set up (Docker)
  - [x] OS installation (PXE boot)
- [x] Automated cluster creation (k3s)
- [x] Automated application deployment (ArgoCD)
- [x] Basic services
  - [x] Gitea
  - [x] Tekton
  - [x] Vault
  - [x] Private container registry
  - [x] Homepage
- [x] Cloudflare tunnel
- [x] Automated DNS management
- [x] Automated certificate management
- [x] Initialize GitOps repository on Gitea automatically
- [x] Observability
  - [x] Monitoring
  - [x] Logging
  - [ ] Alerting
- [ ] Replace all default passwords with randomly generated ones
- [ ] 70% availability (might break in the weekend due to new experimentation)
- [x] Only use open-source technologies (except external managed services in `./external`)
- [x] Everything is defined as code

## Stable requirements

Can be used in "production" (for family or even small scale businesses)

- [x] A single command to deploy everything
- [x] Fast deployment time (from empty hard drive to running services under 1 hour)
- [ ] Fully _automatic_, not just _automated_
  - [ ] Bare-metal OS patching
  - [ ] Backups
  - [ ] Secrets management and rotation
  - [ ] Self healing
- [ ] Additional services
  - [ ] Matrix with bridges
  - [ ] VPN server
  - [ ] PeerTube
  - [x] Seafile
  - [x] Blog
  - [ ] [Development dashboard](https://github.com/khuedoan/homelab-backstage)
- [ ] SSO
- [ ] Backup solution (3 copies, 2 seperate devices, 1 offsite) with encryption
- [ ] 99,9% availability (less than 9 hours of downtime per year)
- [ ] 99,99% data durability
- [ ] Secure by default
  - [ ] SELinux
  - [ ] Network policy
- [ ] Static code analysis
- [ ] Chaos testing
- [ ] Minimal dependency on external services
- [ ] Complete documentation and diagram as code
  - [x] Book (this book)
  - [ ] Walkthrough building tutorial and feature demo (video)
- [ ] Configuration script for new users

## Unplanned

Nice to have

- [ ] Addition services
  - [ ] Mail server
- [ ] Air-gap install
- [ ] Automated testing
- [ ] Security audit
- [ ] Serverless (Knative)
- [ ] Cluster API (https://github.com/khuedoan/homelab/pull/2)
- [ ] Split DNS (requires a better router)
