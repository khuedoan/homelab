# Configuration

## Configure nodes

Edit `infra/production/metal/hosts.json` with your machines' static IPs, MAC addresses, and disk paths.
Installation erases the disks.

!!! example

    ```json title="infra/production/metal/hosts.json"
    {
      "homelab-1": {"ip": "192.168.1.110", "mac_address": "02:00:00:00:00:01", "disk": "/dev/sda"},
      "homelab-2": {"ip": "192.168.1.111", "mac_address": "02:00:00:00:00:02", "disk": "/dev/sda"},
      "homelab-3": {"ip": "192.168.1.112", "mac_address": "02:00:00:00:00:03", "disk": "/dev/sda"},
      "homelab-4": {"ip": "192.168.1.113", "mac_address": "02:00:00:00:00:04", "disk": "/dev/sda"}
    }
    ```

## Configure the cluster

Edit `infra/production/cluster/config.json` with your network settings.
Set `init_host` to a host in the inventory:

```json
{
  "init_host": "homelab-1",
  "vip": "192.168.1.100",
  "load_balancer_cidr": "192.168.1.224/27",
  "network": {"prefix_length": 24, "gateway": "192.168.1.1"}
}
```

## Configure services

Update the following settings in `infra/production/`:

- `domain` in the `cloudflare`, `argocd`, and `platform` Terragrunt units
- `zone_name` in the Cloudflare unit

Review the settings in `apps/*.yaml`, including image registries and homepage links.

After examining the diff, commit and push the changes.
Continue with [deployment](deployment.md).
