# Production metal inventory

`hosts.json` is the active LAN PXE and K3s enrollment inventory. Every host must
have a real, reachable MAC address. Nixie discovers and saves its management IP
during installation; toolbox subsequently enrolls only these hosts.

`deferred-hosts.json` preserves the existing Hetzner machine configuration outside
that automatic path. The remote host has no PXE MAC and cannot be installed by the
LAN broadcast-based Nixie workflow. It is deliberately not installed or enrolled
by `make infra env=production`.

Before activating this remote machine, add and validate an explicit non-PXE
provisioning path, its disk/network configuration, and a private network connecting
it to the cluster API and pod networks. Do not fabricate a MAC or copy it into the
LAN inventory merely to satisfy Nixie's parser. This separation does not remove or
change an already running remote machine.
