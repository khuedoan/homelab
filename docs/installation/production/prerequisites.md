# Prerequisites

## Fork this repository

Because [this project](https://github.com/khuedoan/homelab) applies GitOps practices,
it's the source of truth for _my_ homelab, so you'll need to fork it to make it yours:

[:fontawesome-solid-code-fork: Fork khuedoan/homelab](https://github.com/khuedoan/homelab/fork){ .md-button }

By using this project you agree to [the license](/license).


!!! summary "License TL;DR"

     - This project is free to use for any purpose, but it comes with no warranty
     - You must use the same [GPLv3 license](https://www.gnu.org/licenses/gpl-3.0.en.html)  in `LICENSE.md`
     - You must keep the copy right notice and/or include an acknowledgement
     - Your project must remain open-source

## Hardware requirements

### Initial controller

!!! info

    The initial controller is the machine used to bootstrap the cluster, we only need it once, you can use your laptop or desktop

- A Linux machine that can run Docker (because the `host` networking driver used for PXE boot [only supports Linux](https://docs.docker.com/network/host/), you can use a Linux virtual machine with bridged networking if you're on macOS or Windows).

### Servers

Any modern `x86_64` computer(s) should work, you can use old PCs, laptops or servers.

!!! info

    This is the requirements for _each_ node

| Component  | Minimum                                                                                                      | Recommended                                                                                  |
| :--        | :--                                                                                                          | :--                                                                                          |
| CPU        | 2 cores                                                                                                      | 4 cores                                                                                      |
| RAM        | 8 GB                                                                                                         | 16 GB                                                                                        |
| Hard drive | 128 GB                                                                                                       | 512 GB (depending on your storage usage, the base installation will not use more than 128GB) |
| Node count | 1 (checkout the [single node cluster adjustments](../../how-to-guides/single-node-cluster-adjustments.md) tutorial) | 3 or more for high availability                                                              |

Additional capabilities:

- Ability to boot from the network (PXE boot)
- Wake-on-LAN capability, used to wake the machines up automatically without physically touching the power button

### Network setup

- All servers must be connected to the same **wired** network with the initial controller
- You have the access to change DNS config (on your router or at your domain registrar)

## Domain

Buying a domain is highly recommended, but if you don't have one, see [alternate DNS setup](../../how-to-guides/alternate-dns-setup.md).

## BIOS setup

!!! info

    You need to do it once per machine if the default config is not sufficent,
    usually for consumer hardware this can not be automated
    (it requires something like [IPMI](https://en.wikipedia.org/wiki/Intelligent_Platform_Management_Interface) to automate).

Common settings:

- Enable Wake-on-LAN (WoL) and network boot
- Use UEFI mode and disable CSM (legacy) mode
- Disable secure boot

Boot order options (select one, each has their pros and cons):

1. Only boot from the network if no operating system found: works on most hardware but you need to manually wipe your hard drive or delete the existing boot record for the current OS
2. Prefer booting from the network if turned on via WoL: more convenience but your BIOS must support it, and you must test it throughly to ensure you don't accidentally wipe your servers

!!! example

    Below is my BIOS setup for reference. Your motherboard may have a different name for the options, so you'll need to adapt it to your hardware.

    ```yaml
    Devices:
      NetworkSetup:
        PXEIPv4: true
        PXEIPv6: false
    Advanced:
      CPUSetup:
        VT-d: true
    Power:
      AutomaticPowerOn:
        WoL: Automatic  # Use network boot if Wake-on-LAN
    Security:
      SecureBoot: false
    Startup:
      CSM: false
    ```

## Gather information

- [ ] MAC address for each machine
- [ ] OS disk name (for example `/dev/sda`)
- [ ] Network interface name (for example `eth0`)
- [ ] Choose a static IP address for each machine (just the desired address, we don't set anything up yet)
