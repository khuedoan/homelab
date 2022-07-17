# Dnsmasq

## Overview
dnsmasq is used as either a DHCP server or DHCP proxy server for PXE metal provisioning. Proxy mode can be opted-in through `make configure` or the `metal/roles/pxe_server/defaults/main.yml` file

## When to use Proxy Mode

Proxy mode allows for the project to be deployed to a standard home network where an DHCP server exists, but may not be configurable for next-server or bootfile definitions. The feature can be enabmed through `make confiugure` or modifying the `metal/roles/pxe_server/defaults/main.yml` file. 

It is recommended to set the homelab hosts and their defined IP Addesses (as defined in `metal/inventory/prod.yml`) accordingly on the DHCP server. This will prevent any IP conflicts which may prevent network disruption on your home network. 
