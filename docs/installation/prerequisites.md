# Prerequisites

## Fork this repository

Because [this project](https://github.com/khuedoan/homelab) applies GitOps practices,
it's the source of truth for _my_ homelab, so you'll need to fork it to make it yours:

[Fork khuedoan/homelab](https://github.com/khuedoan/homelab/fork)

## Hardware requirements

### Initial controller

> [!NOTE]
>
> The initial controller is the machine used to bootstrap the cluster, we only need it once, you can use your laptop or desktop

- A **Linux** machine with Nix and `sudo` access (for privileged ports when PXE booting with [Nixie](https://github.com/khuedoan/nixie)).
- An SSH agent, preferably with a TPM-backed key to keep the private key in hardware.
- Add your public SSH key to `infra/src/modules/metal/flake/authorized_keys`.

### Servers

Any modern `x86_64` computer(s) should work, you can use old PCs, laptops or servers.

> [!NOTE]
>
> This is the requirements for _each_ node

| Component | Minimum | Recommended |
| --- | --- | --- |
| CPU | 2 cores | 4 cores |
| RAM | 8 GB | 16 GB |
| Hard drive | 128 GB | 512 GB (depending on your storage usage, the base installation will not use more than 128GB) |
| Node count | 1 (check out the [single node cluster adjustments](../how-to-guides/single-node-cluster-adjustments.md) tutorial) | 3 or more for high availability |

Additional capabilities:

- Ability to boot from the network (PXE boot)
- Wake-on-LAN capability, used to wake the machines up automatically without physically touching the power button

### Network setup

- All servers must be connected to the same **wired** network with the initial controller
- Access to your domain's DNS settings in Cloudflare

## External resources

A domain managed by Cloudflare is required.

Although I try to keep the amount of external resources to the minimum, there's still need for a few of them.
Below is a list of external resources and why we need them (also see some [alternatives](#alternatives) below).

| Provider | Resource | Purpose |
| --- | --- | --- |
| Cloudflare | R2 bucket | OpenTofu state |
| Cloudflare | DNS | DNS and [DNS-01 challenge](https://letsencrypt.org/docs/challenge-types/#dns-01-challenge) for certificates |
| Cloudflare | Tunnel | Public services to the internet without port forwarding |
| Backblaze | B2 bucket | Encrypted volume backups |
| ntfy.sh | Upstream | Push relay for self-hosted ntfy notifications |

To avoid vendor lock-in, each external provider must have an equivalent alternative that is easy to replace:

- Cloudflare R2 and Backblaze B2: any other S3-compatible object storage.
- Cloudflare Tunnel: access everything via VPN, or use port forwarding,
  or a small VPS with WireGuard to route traffic (see also [awesome tunneling](https://github.com/anderspitman/awesome-tunneling)).

### Create credentials

You'll need a password manager to store credentials.
I'm using [`pass`](https://www.passwordstore.org),
but you can use any password manager with a CLI.

Example `infra/${ENV}/.envrc` change for another password manager
(in this case Bitwarden):

```diff
-CLOUDFLARE_API_TOKEN="$(pass "homelab/production/cloudflare/token")"
+CLOUDFLARE_API_TOKEN="$(bw get password "homelab/production/cloudflare/token")"
```

> [!TIP]
> Prefer offline-first password managers such as `pass` or [KeePassXC](https://keepassxc.org) so
> infrastructure bootstrap doesn't depend on an external online service.

#### Cloudflare

1. Add your domain to Cloudflare and enable R2.
2. Create a token under **Manage Account** -> **Account API Tokens** -> **Create Token** -> **Start from scratch**:

| Scope | Permission | Access |
| --- | --- | --- |
| Entire Account | Account API Tokens | Edit |
| Entire Account | Workers R2 Storage | Edit |
| Specified Domains | DNS | Edit |
| Specified Domains | Zone | Read |
| Entire Account | Cloudflare One Connector: cloudflared | Edit |

Then store the credentials in your password manager, e.g.:

```sh
pass insert homelab/production/cloudflare/account-id
pass insert homelab/production/cloudflare/token-id
pass insert homelab/production/cloudflare/token
```

#### Backblaze

Generate a new [B2 master application key](https://secure.backblaze.com/app_keys.htm)
and store the credentials in your password manager, e.g.:

```sh
pass insert homelab/production/backblaze/key-id
pass insert homelab/production/backblaze/key
```

> [!IMPORTANT]
> Each B2 account has one master application key,
> so you likely need to use the same key for both staging and production.

## BIOS setup

> [!NOTE]
>
> You need to do it once per machine if the default config is not sufficient,
> usually for consumer hardware this can not be automated
> (it requires something like [IPMI](https://en.wikipedia.org/wiki/Intelligent_Platform_Management_Interface) to automate).

Common settings:

- Enable Wake-on-LAN (WoL) and network boot
- Use UEFI mode and disable CSM (legacy) mode
- Disable secure boot

Boot order options (select one, each has their pros and cons):

1. Only boot from the network if no operating system found: works on most hardware but you need to manually wipe your hard drive or delete the existing boot record for the current OS
2. Prefer booting from the network if turned on via WoL: more convenience but your BIOS must support it, and you must test it thoroughly to ensure you don't accidentally wipe your servers

> [!NOTE]
> **Example**
>
> Below is my BIOS setup for reference. Your motherboard may have a different name for the options, so you'll need to adapt it to your hardware.
>
> ```yaml
> Devices:
>   NetworkSetup:
>     PXEIPv4: true
>     PXEIPv6: false
> Advanced:
>   CPUSetup:
>     VT-d: true
> Power:
>   AutomaticPowerOn:
>     WoL: Automatic  # Use network boot if Wake-on-LAN
> Security:
>   SecureBoot: false
> Startup:
>   CSM: false
> ```

## Gather information

- [ ] MAC address for each machine
- [ ] OS disk name (for example `/dev/sda`)
- [ ] Choose a static IP address for each machine (just the desired address, we don't set anything up yet)
- [ ] Choose an unused IP range for LoadBalancer Services (e.g. `192.168.1.224/27`) and a control-plane VIP (e.g. `192.168.1.100`)
- [ ] Subnet prefix length and gateway
