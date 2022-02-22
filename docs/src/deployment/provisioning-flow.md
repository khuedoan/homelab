# Provisioning flow

## Overview

![Provisioning flow](../images/provisioning_flow.png)

Everything is automated, after you edit the configuration files, you just need to run a single `make` command and it will:

- Build the `./metal` layer:
  - Create an ephemeral, stateless PXE server
  - Install Linux on all servers in parallel
  - Build a Kubernetes cluster (based on k3s)
- Build the `./bootstrap` layer:
  - Install ArgoCD
  - Install ApplicationSet to manage other layers (and also manage itself)

From now on, ArgoCD will do the rest:

- Build the `./system` layer (storage, networking, monitoring, etc)
- Build the `./platform` layer (Gitea, Vault, SSO, etc)
- Build the `./apps` layer: (Syncthing, Jellyfin, etc)

## Detailed steps

Below is the pseudo code for the entire process, you don't have to read it right now, but it will be handy for debugging.

```
Human run make:
    build ./metal:
        install the OS:
            download the installer image and extract it
            create a PXE server on the controller using Docker Compose:
                DHCP server
                TFTP server
                HTTP server
            create init config for each machine
            turn the machines on via WoL
            the machines boot:
                select network boot automatically
                broadcast DHCP request
                DHCP server reply:
                    machine IP
                    TFTP server (next-server) IP
                get boot files from TFTP server
                    GRUB
                    GRUB config with URL to init config based on MAC address
                    kernel
                    initrd
                boot to the kernel
                download from HTTP server:
                    init config from the URL in GRUB config
                    remaining files required to boot
                install the OS based on the init config:
                    configure the system
                    remaining files required to install
                reboot to the new OS
            controller see the machines are ready
        build a Kubernetes cluster:
            download k3s binary
            generate cluster token
            copy k3s config files
            enable k3s service and form a cluster
            create KUBECONFIG file
            create MetalLB config:
                use the last /27 subnet of the network
                apply the config
    build ./bootstrap:
        install ArgoCD:
            apply helm chart
            wait for status
        install root app:
            select values file:
                if Gitea unreachable (first install):
                    get data from GitHub
                else:
                    get data from Gitea
            apply helm chart
            wait for status
ArgoCD apply the rest:
    clone git repo
    install components based on directories:
        ./bootstrap (it manages itself):
            argocd
            root
        ./system:
            storage
            loadbalancer
            ingress
            etc
        ./platform (depends on ./system):
            git:
                migrate the homelab repository from GitHub
                ArgoCD switch the source from GitHub to Gitea
            ci
            vault
            etc
        ./apps (depends on ./system and ./platform):
            homepage
            jellyfin
            etc
```
