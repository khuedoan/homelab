# Bare-metal

- Ansible renders the configuration file for each bare metal machine (like IP, hostname...) and the PXE server from [templates](./roles/pxe-boot/templates)
- The tools container creates sibling containers to build a PXE server (includes DHCP, TFTP and HTTP server)
- Ansible [wake the machines up](./roles/pxe-boot/tasks/wake.yml) using Wake on LAN
- The machine start the boot process, the OS get installed (through PXE server) and the machine reboots to the new operating system
- Ansible build a Kubernetes cluster based on k3s
