# Hardware requirements

## Initial controller

> The initial controller is the machine used to bootstrap the cluster.

- Any machine that can run Linux and Docker should work (I'm using my laptop).
- Wired Ethernet connection is prefered (Wifi is untested, please let me know if it works)

## Server hardware

> This is the requirements for _each_ node

- Minimum:
  - 1 node
  - At least 2 cores
  - At least 8GB of RAM
  - At least 128GB of hard drive

- Recommended:
  - 3 nodes or more for high availability
  - 4 cores
  - 16GB of RAM
  - 512GB of hard drive (depending on your storage usage, the base installation will not use more than 128GB)

- Ability to boot from the network (PXE boot)
- Wake-on-LAN capability, used to wake the machines up automatically without physically touching the power button
- Connected to the same **wired** network with the initial controller (for DHCP broadcast)
