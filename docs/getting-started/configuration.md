# Configuration

## Configure nodes

Edit `infra/production/metal/hosts.json` with your machines' static IPs, MAC addresses, and disk paths.
Installation erases the disks.

> [!NOTE]
> **Example**
>
> ```json title="infra/production/metal/hosts.json"
> {
>   "homelab-1": {"ip": "192.168.1.110", "mac_address": "02:00:00:00:00:01", "disk": "/dev/sda"},
>   "homelab-2": {"ip": "192.168.1.111", "mac_address": "02:00:00:00:00:02", "disk": "/dev/sda"},
>   "homelab-3": {"ip": "192.168.1.112", "mac_address": "02:00:00:00:00:03", "disk": "/dev/sda"},
>   "homelab-4": {"ip": "192.168.1.113", "mac_address": "02:00:00:00:00:04", "disk": "/dev/sda"}
> }
> ```

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

- `domain` in the `dns`, `tunnel`, `argocd`, and `platform` Terragrunt units
- `zone_name` in the `dns` unit

Review the settings in `apps/*.yaml`, including image registries and homepage links.

## Create credentials

You'll need a password manager to store credentials.
I'm using [`pass`](https://www.passwordstore.org),
but you can use any password manager with a CLI.

Example change to `infra/${ENV}/.envrc` for another password manager
(in this case Bitwarden):

```diff
-CLOUDFLARE_API_TOKEN="$(pass "homelab/production/cloudflare/token")"
+CLOUDFLARE_API_TOKEN="$(bw get password "homelab/production/cloudflare/token")"
```

> [!TIP]
> Prefer offline-first password managers such as `pass` or [KeePassXC](https://keepassxc.org) so
> infrastructure bootstrap doesn't depend on an external online service.

### Cloudflare

1. Add your domain to Cloudflare and enable R2.
2. Create a token under **Manage Account** -> **Account API Tokens** -> **Create Token** -> **Start from scratch**:

| Scope | Permission | Access |
| --- | --- | --- |
| Entire Account | Account API Tokens | Edit |
| Entire Account | Workers R2 Storage | Edit |
| Specified Domains | DNS | Edit |
| Specified Domains | Zone | Read |
| Entire Account | Cloudflare One Connector: cloudflared | Edit |

Then store the credentials in your password manager, e.g.:

```sh
pass insert homelab/production/cloudflare/account-id
pass insert homelab/production/cloudflare/token-id
pass insert homelab/production/cloudflare/token
```

### Backblaze

Generate a new [B2 master application key](https://secure.backblaze.com/app_keys.htm)
and store the credentials in your password manager, e.g.:

```sh
pass insert homelab/production/backblaze/key-id
pass insert homelab/production/backblaze/key
```

> [!IMPORTANT]
> Each B2 account has one master application key,
> so you likely need to use the same key for both staging and production.

---

After examining the diff, commit and push the changes.
Continue with [deployment](deployment.md).
