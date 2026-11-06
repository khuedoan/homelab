# PXE boot

```mermaid
sequenceDiagram
    participant Nixie
    participant Router as Router (DHCP)
    participant Machines

    Nixie->>Nixie: Load hosts.json
    Nixie->>Machines: Check SSH availability
    Note over Nixie: Skip hosts with reachable SSH.<br/>Stop if all hosts are reachable.

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

        Machines->>Nixie: GET /source.tar.gz
        Nixie-->>Machines: Flake source and host configuration
        Machines->>Nixie: POST /status (installation started)
        Note over Machines: Agent builds disko, partitions disks,<br/>and runs nixos-install locally
        Machines->>Nixie: POST /status (installation progress)
        Note over Machines: Schedule reboot into installed system
        Machines->>Nixie: POST /status (installation complete, reboot scheduled)
    end

    Note over Nixie: Shut down after all hosts report completion.<br/>Cluster readiness is checked separately.
```

Here's how it looks like in action (this is an old clip so the logs look different now):

<iframe width="560" height="315" src="https://www.youtube-nocookie.com/embed/y-d7btNNAT8" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
