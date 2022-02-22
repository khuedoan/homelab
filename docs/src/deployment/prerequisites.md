# Prerequisites

## Hardware requirements

### Initial controller

> The initial controller is the machine used to bootstrap the cluster, we only need it once, you can use your laptop or desktop

- Any machine that can run Docker with the `host` networking driver ([which means only Docker on Linux hosts](https://docs.docker.com/network/host/), you can use a Linux virtual machine with bridged networking if you're on macOS or Windows)

### Servers

Any modern `x86_64` computer(s) should work, you can use old PCs, laptops or servers.

> This is the requirements for _each_ node

| Component  | Minimum                                                                                                      | Recommended                                                                                  |
| :--        | :--                                                                                                          | :--                                                                                          |
| CPU        | 2 cores                                                                                                      | 4 cores                                                                                      |
| RAM        | 8 GB                                                                                                         | 16 GB                                                                                        |
| Hard drive | 128 GB                                                                                                       | 512 GB (depending on your storage usage, the base installation will not use more than 128GB) |
| Node count | 1 (checkout the [single node cluster adjustments](../tutorials/single-node-cluster-adjustments.md) tutorial) | 3 or more for high availability                                                              |

Additional capabilities:

- Ability to boot from the network (PXE boot)
- Wake-on-LAN capability, used to wake the machines up automatically without physically touching the power button

### Network setup

- All servers must be connected to the same **wired** network with the initial controller (Wifi is untested, please let me know if it works)
- You have the access to change DNS config (on your router or at your domain registrar)

## Domain

Buying a domain is highly recommended, but if you don't have one, you can also update your router config and point [`*.home.arpa`](https://datatracker.ietf.org/doc/html/rfc8375) to the load balancer (more on that later).

## BIOS setup

> You need to do it once per machine if the default config is not sufficent,
> usually for consumer hardware this can not be automated
> (it requires something like [IPMI](https://en.wikipedia.org/wiki/Intelligent_Platform_Management_Interface) to automate).

Common settings:

- Enable Wake-on-LAN (WoL) and network boot
- Use UEFI mode and disable CSM (legacy) mode
- Disable secure boot

Boot order options (select one, each has their pros and cons):

1. Only boot from the network if no operating system found: works on most hardware but you need to manually wipe your hard drive or delete the existing boot record for the current OS
2. Prefer booting from the network if turned on via WoL: more convenience but your BIOS must support it, and you must test it throughly to ensure you don't accidentally wipe your servers

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
