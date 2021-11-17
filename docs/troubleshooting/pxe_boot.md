# PXE boot

## PXE server logs

To view PXE server (includes DHCP, TFTP and HTTP server) logs:

```sh
./scripts/pxe-logs
```

## Nodes not booting from the network

- Plug a monitor and a keyboard to one of the bare metal node if possible to make the debugging process easier
- Check if the controller (PXE server) is on the same subnet with bare metal nodes (sometimes Wifi will not work or conflict with wired Ethernet, try to turn it off)
- Check if bare metal nodes are configured to boot from the network
- Check if Wake-on-LAN is enabled
- Check if the operating system ISO file is mounted
- Check the controller firewall config
- Check PXE server Docker logs
