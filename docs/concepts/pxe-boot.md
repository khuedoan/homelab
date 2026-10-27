# PXE boot

```mermaid
sequenceDiagram
    participant Nixie
    participant Router as Router (DHCP)
    participant Machines

    Nixie->>Nixie: Load hosts.json
    Nixie->>Machines: Check recorded machine IDs over SSH
    Note over Nixie: Skip recorded hosts, even if unreachable<br/>or their identity differs. Stop if all are recorded.

    Nixie->>Nixie: Build NixOS installer with Nix
    Nixie->>Nixie: Start ProxyDHCP, TFTP, HTTP, and API services
    loop Each uninstalled machine
        Nixie->>Machines: Send Wake-on-LAN packet
    end

    par Machines boot and install concurrently
        Note over Machines: UEFI PXE boot
        Machines->>Router: Broadcast DHCP discovery
        Machines->>Nixie: Same broadcast
        par IP configuration
            Router-->>Machines: Assign IP address
        and PXE discovery
            Nixie-->>Machines: Offer PXE boot server (ProxyDHCP)
        end
        Machines->>Nixie: Discover boot file (UDP 4011)
        Nixie-->>Machines: iPXE bootloader path
        Machines->>Nixie: Request bootloader (TFTP)
        Nixie-->>Machines: iPXE bootloader

        Note over Router,Machines: iPXE repeats DHCP discovery<br/>ProxyDHCP supplies the boot-script URL
        Machines->>Nixie: Request boot script, then kernel and initrd (HTTP)
        Nixie-->>Machines: Boot script, kernel, and initrd
        Note over Machines: Boot NixOS installer<br/>systemd starts nixie-agent

        Machines->>Nixie: POST /install with MAC address
        Nixie->>Nixie: Select host flake by MAC<br/>Use request's source IP for SSH
        Nixie->>Machines: Run nixos-anywhere to install NixOS over SSH
        Note over Machines: Reboot into installed system
        Nixie->>Machines: Read installed /etc/machine-id over SSH
        Machines-->>Nixie: Machine ID
        Nixie->>Nixie: Save IP and machine ID hash to hosts.json<br/>Mark host installed
    end

    Note over Nixie: Shut down after all hosts are installed
```

Here's how it looks like in action (this is an old clip so the logs look different now):

<iframe width="560" height="315" src="https://www.youtube-nocookie.com/embed/y-d7btNNAT8" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
