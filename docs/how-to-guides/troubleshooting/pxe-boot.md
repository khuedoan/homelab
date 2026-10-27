# PXE boot

## PXE server logs

View the Nixie output in the terminal running `make infra env=production` (or staging).
To retry metal provisioning:

```sh
cd infra/production/metal
terragrunt run apply
```

## Nodes not booting from the network

- Plug a monitor and a keyboard to one of the bare metal node if possible to make the debugging process easier
- Check if the controller (PXE server) is on the same subnet with bare metal nodes (sometimes Wifi will not work or conflict with wired Ethernet, try to turn it off)
- Check if bare metal nodes are configured to boot from the network
- Check if Wake-on-LAN is enabled
- Check the controller's firewall config to make sure that the following ports are open:
    - UDP 67 and 68 for DHCP
    - UDP 69 for TFTP
    - UDP 4011 for the PXE boot service
    - TCP 80 for the boot files
    - TCP 5000 for the installer callback API
- Check if the servers are booting to the correct OS (NixOS installer instead of the previously installed OS), if not try to select it manually or remove the previous OS boot entry
- Examine the network boot process with [Wireshark](https://www.wireshark.org) or [Termshark](https://termshark.io)

Verify your authorized SSH keys and trusted host keys if provisioning succeeds but enrollment fails.
