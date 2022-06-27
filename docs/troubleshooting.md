# Troubleshooting

## PXE server logs

To view PXE server (includes DHCP, TFTP and HTTP server) logs:

```sh
./scripts/pxe-logs
```

You can view the logs of one or more containers selectively, for example:

```sh
./scripts/pxe-logs dhcp
./scripts/pxe-logs tftp http
```

## Nodes not booting from the network

- Plug a monitor and a keyboard to one of the bare metal node if possible to make the debugging process easier
- Check if the controller (PXE server) is on the same subnet with bare metal nodes (sometimes Wifi will not work or conflict with wired Ethernet, try to turn it off)
- Check if bare metal nodes are configured to boot from the network
- Check if Wake-on-LAN is enabled
- Check if the operating system ISO file is mounted
- Check the controller's firewall config
- Check PXE server Docker logs
- Check if the servers are booting to the correct OS (Rocky Linux installer instead of the previously installed OS), if not try to select it manually or remove the previous OS boot entry
