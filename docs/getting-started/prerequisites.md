# Prerequisites

## Fork this repository

Because [this project](https://github.com/khuedoan/homelab) applies GitOps practices,
it's the source of truth for _my_ homelab, so you'll need to fork it to make it yours:

[Fork khuedoan/homelab](https://github.com/khuedoan/homelab/fork)

## Hardware requirements

### Initial controller

> [!NOTE]
>
> The initial controller is the machine used to bootstrap the cluster. We only need it once; you can use your laptop or desktop.

- A **Linux** machine with Nix and `sudo` access (for privileged ports when PXE booting with [Nixie](https://github.com/khuedoan/nixie)).
- An SSH agent, preferably with a TPM-backed key to keep the private key in hardware.
- Add your public SSH key to `infra/src/modules/metal/flake/authorized_keys`.

### Servers

Any modern `x86_64` computer should work. You can use old PCs, laptops, or servers.

> [!NOTE]
>
> These are the requirements for _each_ node.

| Component | Minimum | Recommended |
| --- | --- | --- |
| CPU | 2 cores | 4 cores |
| RAM | 8 GB | 16 GB |
| Hard drive | 128 GB | 512 GB (depending on your storage usage; the base installation will not use more than 128 GB) |
| Node count | 1 (check out the [single node cluster adjustments](../how-to-guides/single-node-cluster-adjustments.md) tutorial) | 3 or more for high availability |

Additional capabilities:

- Ability to boot from the network (PXE boot)
- Wake-on-LAN capability, used to wake the machines up automatically without physically touching the power button

### Network setup

- All servers must be connected to the same **wired** network as the initial controller
- Access to your domain's DNS settings in Cloudflare

## BIOS setup

> [!NOTE]
>
> You need to do this once per machine if the default config is not sufficient.
> Usually, this cannot be automated on consumer hardware
> (it requires something like [IPMI](https://en.wikipedia.org/wiki/Intelligent_Platform_Management_Interface) to automate).

Common settings:

- Enable Wake-on-LAN (WoL) and network boot
- Use UEFI mode and disable CSM (legacy) mode
- Disable secure boot

Boot order options (select one; each has its pros and cons):

1. Only boot from the network if no operating system is found: works on most hardware but you need to manually wipe your hard drive or delete the existing boot record for the current OS
2. Prefer booting from the network if turned on via WoL: more convenient, but your BIOS must support it, and you must test it thoroughly to ensure you don't accidentally wipe your servers

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

## External resources

Although I try to keep external resources to a minimum, a few are still needed.
Create an account for:

- Cloudflare
- Backblaze

We use them for:

| Provider | Resource | Purpose |
| --- | --- | --- |
| Cloudflare | R2 bucket | OpenTofu state |
| Cloudflare | DNS | DNS and [DNS-01 challenge](https://letsencrypt.org/docs/challenge-types/#dns-01-challenge) for certificates. A domain managed by Cloudflare is required. |
| Cloudflare | Tunnel | Expose services to the internet without port forwarding |
| Backblaze | B2 bucket | Encrypted volume backups |

To avoid vendor lock-in, each external provider must have an equivalent alternative that is easy to replace:

- Cloudflare R2 and Backblaze B2: any other S3-compatible object storage.
- Cloudflare Tunnel: access everything via VPN, or use port forwarding,
  or a small VPS with WireGuard to route traffic (see also [awesome tunneling](https://github.com/anderspitman/awesome-tunneling)).
