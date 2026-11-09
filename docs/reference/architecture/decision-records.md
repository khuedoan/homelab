# Decision records

These are the records of design decisions for future reference in order to understand why things are the way they are.
They are not permanent, I can change them in the future if better alternatives become available.

<details>
<summary>Template</summary>

## Description of the change

**Context**

CHANGEME

**Decision**

CHANGEME

**Consequences**

- CHANGEME

</details>

## Current decisions

### Using NixOS as the base OS

**Context**

With Fedora and Ansible, I sometimes make manual changes and forget to put them back in the playbooks.
If a playbook fails halfway, it's also hard to roll back to the previous system.

I've also considered other atomic distros, for example Talos is smaller and has better integration for Kubernetes operations,
but I want a general-purpose OS that can stay even if I move away from Kubernetes.
I don't want the OS to depend on a single company that could change its license,
and I already use NixOS on my other machines.

**Decision**

Use NixOS as the base OS.

**Consequences**

- System changes have to go through the NixOS configuration, so it's easier to keep the system clean and reproducible.
- NixOS builds the new system before switching to it, and I can roll back to a previous generation.
  Switching to the new system is almost fully atomic, except for service activation and restarts, which are still imperative.
- I can use Nix for both the OS and development shells, without installing project dependencies into the base system.
- I can test the same configurations in QEMU before applying them to real hardware.
- I can keep the OS even if I replace Kubernetes later.
- The system is larger than Talos, with a larger attack surface, and I need to coordinate rolling upgrades myself.

See the [original discussion](https://github.com/khuedoan/homelab/pull/197#issuecomment-3440096206)
and [NixOS cured my OCD](https://www.khuedoan.com/posts/nixos-cured-my-ocd) for more background.

### Switch metal provisioning from Ansible to Nixie

**Context**

The metal nodes used to use Ansible to install Fedora.
When I switched to NixOS, I could still use Ansible to install it,
but it was hard to test and debug.

**Decision**

Build a custom bare metal orchestration engine for NixOS - [nixie](https://github.com/khuedoan/nixie).

Nixie installs from the NixOS configuration, replacing Kickstart without needing cloud-init.
It works with VMs too. My staging cluster runs on Incus VMs that I regularly destroy and recreate,
while production runs on bare metal.
See [PXE boot](../../concepts/pxe-boot.md) for the installation flow.

**Consequences**

- Single binary with native OS process instead of complex Ansible & Docker Compose setup (& its host networking baggage)
- It has much better integration to the OS, since it's built specically for NixOS
- It's much more reliable and testable (Nixie has a full test suite in QEMU)
- There's much better visibility into the PXE boot process with tracing
- The PXE installer can be reused with other projects
  (e.g., my [dotfiles](https://github.com/khuedoan/dotfiles)),
  not just this homelab.

### K3s

K3s is reliable, well supported, and available in `nixpkgs` with a NixOS module.
I've been using it for years without issues that would make me want to replace it.
I replace the bundled Traefik, ServiceLB, and local storage with Istio, kube-vip, and Rook Ceph.

### Replace ingress-nginx with Istio

I want a service mesh for observability, security, and traffic routing, and I need Gateway API support anyway.
Istio provides both, and I'm already running it in production.
Ambient mesh has the features I need without sidecar overhead.

I keep Flannel for pod networking and kube-vip for the control-plane VIP and Service load balancing.
Keeping the existing CNI makes it easy to port the platform to cloud-managed Kubernetes without bringing my own CNI.

Application routes use Gateway API, and ambient mesh encrypts traffic between application Pods with strict mTLS.
See [Networking](networking.md) for the current setup.

### Use Rook Ceph for storage

Ceph is reliable, battle-tested distributed storage, and Rook manages it in Kubernetes.
I've tried Longhorn but still ran into volume detachment issues.
A single NFS server would be a single point of failure, which I don't want for cluster storage.
Ceph is more complex to operate, but its reliability is worth it to me.

### Use Velero for backups

**Context**

Requirements:

- All apps backed up by default, without additional configuration for new apps.
- The backups need to work even if I lose the entire cluster.

I used to use [VolSync](https://volsync.readthedocs.io/en/stable/usage/restic/index.html),
but its Restic workflow requires per-volume replication resources and credentials,
and the backup & restore flow is not really declarative.
I would need additional resource generation and orchestration to provide this experience.

**Decision**

Use Velero to back up cluster resources and supported persistent volumes:

- Use [CSI snapshot data movement](https://velero.io/docs/v1.18/csi-snapshot-data-movement)
  to copy volume data into object storage,
  so recovery does not depend on the original Ceph cluster.
- Use Velero's [ReadOnly recovery mode](https://velero.io/docs/v1.18/disaster-case)
  to prevent backup creation and deletion while restoring and verifying data.

**Consequences**

- The backup schedule includes all apps,
  without backup resources or storage credentials to distribute to app namespaces.
- Backup and restore operations use Velero's native resources and CLI.
- Velero requires broad cluster permissions.
- Database applications may still need backup hooks or native exports for consistent backups.

### Use Kanidm for identity

I want a reliable, secure identity provider.
I trust Kanidm's maintainer and track record, and its use of Rust gives me more peace of mind.

I only use Kanidm and don't plan to switch identity providers.
Dex helped with client IDs and secrets, but added another layer and limited which Kanidm features I could use.
I now integrate with Kanidm directly and automate client provisioning myself.
If I switch providers later, rewriting the automation scripts is a small enough cost.

### Use OpenBao for secrets

I need an API to provision application credentials automatically.
For example, SSO setup can create a Kanidm client secret and write it to OpenBao for the application to use.

I could do this with the Kubernetes Secret API, but I also want secret injection and namespace-scoped access policies.
Applications should only be able to read secrets under their own namespace's path.
OpenBao provides those guardrails, with Bank-Vaults handling secret injection.
It's more to maintain than the previous global secret store, but worth it for these features.

See [Secrets management](../../concepts/secrets-management.md) for the implementation.

### Manage platform components through OpenTofu and ArgoCD

I already need OpenTofu for infrastructure and a way to bootstrap the platform.
It gives me templating and dependency ordering, while ArgoCD deploys and reconciles the workloads.
I prefer HCL to Helm templates or string replacement.
User applications stay in GitOps, managed by ArgoCD.

Terragrunt gives each provisioning unit its own state and provider configuration.
It handles dependencies between units without custom orchestration.
The cluster must exist before I can initialize the Kubernetes and Helm providers for platform components.

OpenTofu can also provision secrets before OpenBao is available.
Velero needs this because it must be able to restore OpenBao, so its credentials can't depend on OpenBao being up first.

### Use app-template for all applications

I want consistent application configuration so I can apply platform requirements like ambient mesh and future security settings in one place.
All applications use [app-template](https://bjw-s-labs.github.io/helm-charts/docs/app-template), with a shared ApplicationSet handling platform integration.
I lose some flexibility and can't reuse upstream Helm charts directly, but so far I've been able to convert every application.

App-template supports arbitrary resources through `rawResources`, so the chart alone isn't a guardrail.
I use separate ArgoCD projects: platform components need cluster-wide privileges, while applications have a resource allowlist.
This gives me more confidence that application changes have a contained blast radius.

### Use Forgejo as the GitOps source

The point of this project is to self-host, so that includes Git.
I don't want a critical platform feature to depend on a closed-source third-party service like GitHub.
Forgejo lets me control the GitOps source end-to-end.
If it's down, I can still update things manually or fix it myself.

Self-hosting everything has to be practical. If it's too complex, people won't do it.
I'm fine with managed services that provide enough value without much lock-in.
DNS and S3-compatible storage are good examples: they're commodity services, and I still own my data.
Tunneling is another place where a managed service can be worth it, as long as I can replace it.

### Keep Tailscale as a fallback for WireGuard

I primarily use WireGuard, but my ISP has broken access through my public IP several times.
Tailscale gives me another way in when that happens.
I don't want to depend on its managed control plane as my primary access method, so I keep it as a fallback.

### Remove the Docker wrapper for Nix shell

**Context**

The Docker wrapper was originally added because not everyone has Nix installed.
However, it is not a good abstraction. Depending on how Docker (or Podman) was
installed, there may be permission issues, and it is slower than running Nix
directly on the host.

**Decision**

Remove the Docker wrapper and run the Nix development shell directly. While not
everyone has Nix installed, not everyone has the same Docker setup either.

**Consequences**

- Requires an additional step to install Nix, but this is a one-time setup.
- The `make tools` command has been removed, you now need to run `nix develop`
  directly. Although it can be added to the `Makefile`, this requires `make` to
  be installed, and not everyone has the same version of `make`.

### Replace Make with Just

Most of my commands are things I need to run, rather than things I need to build.
Just fits this better, with positional arguments and support for Python recipes.
Modules let me keep recipes alongside the infrastructure and tests, exposed through the root `justfile`.
It needs an extra installation, but I already have a Nix shell that provides it.

## Historical decisions

These records describe earlier choices and are kept for context.

### Remove HashiCorp Vault

**Status**: superseded by [Use OpenBao for secrets](#use-openbao-for-secrets), the community fork of Vault.

**Context**

- HashiCorp changed their license, and it's no longer free/libre software.
  One of the highest priorities of this project is to minimize
  the usage of non-free software as much as possible, so I don't really
  want to keep Vault, especially considering the next point.
- Vault is fairly complex to maintain properly. This project only uses
  Vault for two things: basic key-value secret store and its API to
  create and manage secrets dynamically. With the new Kubernetes secret
  provider in External Secrets, both features can be replaced with
  Kubernetes's built-in secrets and API server.
- A related goal of using Vault as an identity provider for SSO will be
  discarded, and I'll use Authelia instead, which has a beta identity
  provider feature (or use another alternative).

**Decision**

Replace Vault with a simpler in-cluster global secret store.

**Consequences**

Unlike secret path in Vault, Kubernetes does not support `/` in object name.
I need to change secret convention from `path` to `name` and replace `/` with `.`.

Update secret generator config:

```diff
-- path: gitea/admin
+- name: gitea.admin
   data:
     - key: password
       length: 32
```

Update secret references in `kind: ExternalSecret`:

```diff
 remoteRef:
-  key: /gitea/admin
+  key: gitea.admin
```

### Manage package versions in development shell

**Status**: I no longer need this, since I use Nix from the development shell to NixOS on the nodes.

**Context**

While Nix is reproducible, I need a way to control the versions of the tools and keep them up-to-date.
For example, if I update the nixpkgs hash (in `flake.nix`) from `abcd1234` to `defa5678`:

- `ansible`: 2.12.1 -> 2.12.6
- `terraform`: 1.2.0 -> 1.2.2
- `foobar`: 1.8.0 -> 1.9.0

That looks good. But when I update it from `defa5678` to `cdef9012`:

- `ansible`: 2.12.6 -> 2.13.0
- `terraform`: 1.2.2 -> 1.3.1
- `foobar`: 1.9.0 -> 2.0.0

This time it breaks `foobar` because the new major version contains a breaking change.

I can pin the specific version of each dependency in `flake.nix`,
however, the maintenance burden is too high (even with Renovate) because I need to update the version of each package regularly rather than just the nixpkgs hash.
Instead, I can just bump the nixpkgs hash and run some tests to ensure there is no breaking change.

**Decision**

Update the tests to ensure that the versions remain within the desired range (i.e. no breaking change).

**Consequences**

I have the rail guard from the tests to ensure that I don't upgrade to a new major version with breaking changes,
and I can make a conscious decision to take the necessary steps prior to upgrading to the new major version.

### Refactor the tools container from plain Dockerfile to [Nix](https://nixos.org)

**Status**: the Docker wrapper was later removed in [Remove the Docker wrapper for Nix shell](#remove-the-docker-wrapper-for-nix-shell).

**Context**

The previous implementation of the tools container is not reproducible, if someone builds it a few weeks after me, they will receive different versions of the tools.
Also, if you change something in the tool list, everything will be downloaded again, with no caching.

**Decision**

Move to Nix shell with a Docker wrapper, in case Nix is not available (see commit `adbaf32`).

**Consequences**

- All versions are pinned
- When you change the list of tools, rebuilding is much faster

### Combine dhcpd and tftpd to dnsmasq in PXE server

**Status**: replaced by [Nixie](#switch-metal-provisioning-from-ansible-to-nixie).

**Context**

[Original proposal from @Bourne-ID](https://github.com/khuedoan/homelab/issues/70):

> **Issue statement**
>
> The use of dhcpd is great for air-gap solutions where a new DHCP is required. However for some home networks which does not have the VLAN capability or for users who would like to use common router DHCP services, the use of DHCPD will cause duplicate DHCP servers and will result in potential network disruption, or will limit the ability to auto-provision the Metal stage of this project.
>
> **Proposed Solution: DHCP Proxy**
>
> Use DHCP Proxy services to add PXE features such as Next Server into this project. This allows for users to use the existing DHCP servers which may be locked down or incapable of using Next Server/PXE settings on their network to be able to auto-provision hardware through PXE (with certain common configurations, like static IP allocation or reduction in DHCP request ranges on the DHCP server)
>
> **Proposed Application: DNSMasq**
>
> DNSMasq in Proxy mode interoperates with existing DHCP servers over IPv4 to add features such as next-server, TFTP, etc. where such hardware is either locked or unconfigurable for such services. This would be an opt-in change, configurable through the pxe_server defaults file.
>
> **Proposed Target Audience**
>
> Users who either do not want to create their own VLAN or lack the hardware to configure such services. Users who want to use common router services for DHCP and have router access to configure static IP and/or DHCP allocation ranges.
>
> **Additional Risks with Proposed Change**
>
>   - Additional Surface Area for Break-Out Attacks: Originally this project is locked to its own DHCP/VLAN, so any break-outs should be contained accordingly. Using common home networks increases the surface area of break-out attacks if the deployment is compromised.
>   - Mitigation: Enrolment into this change is opt in only.
>
> **Proposed Next Steps**
>
> 1. Trial/Adopt/Halt - A discussion with all or a decision by the project maintainers to identify if this change should exist in this project or live on a fork.
> 2. Documentation (This is in flight in any situation).

**Decision**

Migrate to dnsmasq (see commit `f650c89`, thanks to [@Bourne-ID](https://github.com/Bourne-ID))

**Consequences**

- DHCP proxy is enabled by default because most people have a standard home network with existing DHCP server, but it can still be disabled to restore the previous behavior

### Migrate documentation from [mdBook](https://rust-lang.github.io/mdBook) to [MkDocs](https://squidfunk.github.io/mkdocs-material)

**Status**: I no longer plan to use Backstage for this homelab (I may build something else),
so it's back to mdBook, which has fewer dependencies and is faster.

**Context**

mdBook is very minimal and light weight, which I personally prefer.
However, [Backstage TechDocs](https://backstage.io/docs/features/techdocs/techdocs-overview) which I plan to use currently only supports mkDocs.

**Decision**

Migrate documentation from mdBook to MkDocs (see commit `cd41343`).

**Consequences**

- I can no longer include only a portion of a file, see [facelessuser/pymdown-extensions#1462](https://github.com/facelessuser/pymdown-extensions/issues/1462).

### Using Fedora as the base OS

**Status**: superseded by [Using NixOS as the base OS](#using-nixos-as-the-base-os)

**Context**

I've tried several distributions, and each has advantages and disadvantages.
Fedora has a good (enough) balance between stability and new features.

Alternatives considered:

- Fedora CoreOS (moved to Rocky in `7667254`):
    - Pros: automatic and atomic upgrade, immutable, quick installation
    - Cons: hard to run Ansible on (Python is not included)
- CentOS/Rocky Linux (moved to Fedora in `022b816`):
    - Pros: relatively stable (however I did encounter a breaking change [#63](https://github.com/khuedoan/homelab/issues/63), still not sure why)
    - Cons: kernel and packages are too old
- Debian: couldn't get it to work with PXE boot and Rocky Linux was sufficient so I didn't push any further
- Cluster API (previous attempt in ` a8e4a85`, I hope to get this to work someday):
    - Pros: control bare metal machines via Kubernetes API, open the possibility for autoscaling and autohealing
    - Cons: doesn't support simple WoL and shutdown via SSH (or similar)

**Decision**

Use Fedora as the base OS.

**Consequences**

`¯\_(ツ)_/¯`
