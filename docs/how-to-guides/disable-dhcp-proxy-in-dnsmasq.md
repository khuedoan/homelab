# Disable DHCP proxy in dnsmasq

## Overview

Dnsmasq is used as either a DHCP server or DHCP proxy server for PXE metal provisioning.

Proxy mode is enabled by default allowing the use of existing DHCP servers on the network.
A good description on how DHCP Proxy works can be found on the related [FOG project wiki page](https://wiki.fogproject.org/wiki/index.php?title=ProxyDHCP_with_dnsmasq).

## Disabling Proxy mode

Certain scenarios will require this project to use a DHCP server, such as an air-gap deployment or dedicated VLAN.

To disable proxy mode thereby using dnsmasq as a DHCP server, modify `metal/roles/pxe_server/defaults/main.yml` and set `dhcp_proxy` to `false`.
