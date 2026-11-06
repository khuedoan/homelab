# Roadmap

> [!NOTE]
>
> Current status: **BETA**

## Alpha requirements

Literally anything that works.

## Beta requirements

Good enough for tinkering and personal use, and reasonably secure.

- [x] Automated bare metal provisioning
    - [x] Controller setup (Nix development shell)
    - [x] OS installation (PXE boot)
- [x] Automated cluster creation (k3s)
- [x] Automated application deployment (ArgoCD)
- [x] Automated DNS management
- [x] Initialize the GitOps repository on Forgejo automatically
- [x] Observability
    - [x] Monitoring
    - [x] Logging
    - [x] Alerting
- [x] SSO
- [x] Reasonably secure
    - [x] Automated certificate management
    - [x] Declarative secret management
    - [x] Replace all default passwords with randomly generated ones
    - [x] Expose services to the internet securely with Cloudflare Tunnel
- [x] Only use open-source technologies (except external managed services)
- [x] Everything is defined as code
- [ ] Backup solution (3 copies, 2 separate devices, 1 offsite)
    - [x] Scheduled, encrypted offsite backups
    - [ ] Verify the full 3-2-1 backup strategy
- [ ] Define [SLOs](https://en.wikipedia.org/wiki/Service-level_objective):
    - [ ] 99% availability per application (up to 7h of downtime per month, including weekend experimentation)
- [x] Core applications
    - [x] Forgejo
    - [x] Woodpecker
    - [x] Private container registry
    - [x] Homepage

## Stable requirements

Can be used in "production" (for families or even small businesses).

- [x] A single command to deploy everything
- [ ] Verify deployment time from empty disks to running services
- [ ] Fully _automatic_, not just _automated_
    - [ ] Automated flake updates and node configuration deployment for OS and Kubernetes upgrades
        - [ ] Rolling upgrades with cordon/drain, control-plane/worker coordination, and health checks
    - [x] Kured reboot coordination after kernel changes
    - [ ] Application version upgrades
    - [x] Encrypted backups
    - [ ] Secret rotation and consumer reloads
    - [x] Self-healing
- [ ] Secure by default
    - [ ] Host security hardening
    - [ ] Default application network policies
- [x] Static code analysis
- [x] Automated integration tests (networking, storage, backup/restore, secrets, GitOps, etc.)
- [ ] Regular disaster recovery tests
- [x] Minimal dependency on external services
- [ ] Complete documentation
    - [x] Diagrams as code
    - [x] Book (this book)
    - [ ] More tutorials and video feature demos
- [x] User onboarding (with `just user`)
- [ ] More dashboards and alert rules
- [ ] Monitoring and alert delivery outside the cluster
- [ ] Verify mobile/desktop app clients and off-LAN VPN access
- [ ] Capacity planning and control-plane stability under load
- [ ] SLOs:
    - [ ] 99.9% availability (up to 43 mins of downtime per month). Ideally 99.95% but would be hard to do with home power and internet (I might move to colo)
    - [ ] Recovery from complete cluster loss (at most one hour of data loss)
- [ ] Clear upgrade path
- [ ] Additional applications
    - [ ] Matrix with bridges
    - [x] VPN server
    - [ ] PeerTube
    - [x] Blog

## Unplanned

Nice to have.

- [ ] Air-gapped installation
- [ ] Chaos testing
