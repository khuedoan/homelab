# Decision records

These are the records of design decisions for future reference in order to understand why things are the way they are.
They are not permanent, we can change them in the future if better alternatives become available.

??? Template

    ## Description of the the change

    **Context**

    CHANGEME

    **Decision**

    CHANGEME

    **Consequences**

    - CHANGEME

## Remove the Docker wrapper for Nix shell

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

## Remove HashiCorp Vault

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
  discarded, and we'll use Authelia instead, which has a beta identity
  provider feature (or use another alternative).

**Decision**

Replace Vault with a simplier in-cluster global secret store.

**Consequences**

Unlike secret path in Vault, Kubernetes does not support `/` in object name.
We need to change secret convention from `path` to `name` and replace `/` with `.`.

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

## Manage package versions in development shell

**Context**

While Nix is reproducible, we need a way to control the versions of the tools and keep them up-to-date.
For example, if we update the nixpkgs hash (in `flake.nix`) from `abcd1234` to `defa5678`:

- `ansible`: 2.12.1 -> 2.12.6
- `terraform`: 1.2.0 -> 1.2.2
- `foobar`: 1.8.0 -> 1.9.0

That looks good. But when we update it from `defa5678` to `cdef9012`:

- `ansible`: 2.12.6 -> 2.13.0
- `terraform`: 1.2.2 -> 1.3.1
- `foobar`: 1.9.0 -> 2.0.0

This time it breaks `foobar` because the new major version contains a breaking change.

We can pin the specific version of each dependency in `flake.nix`,
however, the maintenance burden is too high (even with Renovate) because we need to update the version of each package regularly rather than just the nixpkgs hash.
Instead, we can just bump the nixpkgs hash and run some tests to ensure there is no breaking change.

**Decision**

Update the tests to ensure that the versions remain within the desired range (i.e. no breaking change).

**Consequences**

We have the rail guard from the tests to ensure that we don't upgrade to a new major version with breaking changes,
and we can make a conscious decision to take the necessary steps prior to upgrading to the new major version.

## Refactor the tools container from plain Dockerfile to [Nix](https://nixos.org)

**Context**

The previous implementation of the tools container is not reproducible, if someone builds it a few weeks after me, they will receive different versions of the tools.
Also, if you change something in the tool list, everything will be downloaded again, with no caching.

**Decision**

Move to Nix shell with a Docker wrapper, in case Nix is not available (see commit `adbaf32`).

**Consequences**

- All versions are pinned
- When you change the list of tools, rebuilding is much faster

## Combine dhcpd and tftpd to dnsmasq in PXE server

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

## Migrate documentation from [mdBook](https://rust-lang.github.io/mdBook) to [MkDocs](https://squidfunk.github.io/mkdocs-material)

**Context**

mdBook is very minimal and light weight, which I personally prefer.
However, [Backstage TechDocs](https://backstage.io/docs/features/techdocs/techdocs-overview) which I plan to use currently only supports mkDocs.

**Decision**

Migrate documentation from mdBook to MkDocs (see commit `cd41343`).

**Consequences**

- We can no longer include only a portion of a file, see [facelessuser/pymdown-extensions#1462](https://github.com/facelessuser/pymdown-extensions/issues/1462).

## Choosing the base OS

**Context**

I've tried several distributions, and each has advantages and disadvantages.
Fedora has a good (enough) balance between stability and new features.

Alternatives considered:

- Fedora CoreOS (moved to Rocky in `7667254`):
    - Pros: automatic and atomic upgrade, immutable, quick installation
    - Cons: hard to run Ansible on (Python is not included)
- CentOS/Rocky Linux (moved to Fedora in `022b816`):
    - Pros: relatively stable (however we did encounter a breaking change [#63](https://github.com/khuedoan/homelab/issues/63), still not sure why)
    - Cons: kernel and packages are too old
- Debian: couldn't get it to work with PXE boot and Rocky Linux was sufficient so I didn't push any further
- Cluster API (previous attempt in ` a8e4a85`, I hope to get this to work someday):
    - Pros: control bare metal machines via Kubernetes API, open the possibility for autoscaling and autohealing
    - Cons: doesn't support simple WoL and shutdown via SSH (or similar)

**Decision**

Use Fedora as the base OS.

**Consequences**

`¯\_(ツ)_/¯`
