# VPN setup

You can choose between [Tailscale](https://tailscale.com), [Wireguard](https://www.wireguard.com), or use both like me.

## Tailscale (requires third-party account)

Get an [auth key](https://tailscale.com/kb/1085/auth-keys) from [Tailscale admin console](https://login.tailscale.com/admin/authkeys):

- Description: homelab
- Reusable: optionally set this to true

Add it to `external/terraform.tfvars` as an extra secret:

```hcl
extra_secrets = {
  tailscale-auth-key = "tskey-auth-myauthkeyhere"
}
```

Apply the secret:

```sh
make external
```

Finally, [enable subnet routes](https://tailscale.com/kb/1019/subnets#step-3-enable-subnet-routes-from-the-admin-console) for `homelab-router`
from the [admin console](https://login.tailscale.com/admin/machines).

You can now connect to your homelab via Tailscale and [invite user to your Tailscale network](https://tailscale.com/kb/1371/invite-users).

## Wireguard (requires port-forwarding)

Update the peer list in `apps/wireguard/values.yaml`:

```yaml
PEERS: |
  UserDevice
  FooPhone
  FooLaptop
  BarDesktop
```

Go to your router settings and forward the Wireguard service.
Each router is different, here's mine for reference:

- Protocol: `UDP`
- Start Port: `51820`
- End Port: `51820`
- Local IP Address: `192.168.1.226` (find it with `kubectl get service -n wireguard wireguard`)
- Start Port Local: `51820`
- End Port Local: `51820`

To get the QR code (for mobile) and config (for desktop), run:

!!! warning

    This command will print sensitive secrets to the terminal.

```sh
./scripts/get-wireguard-config FooPhone
```
