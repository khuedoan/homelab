# VPN setup

You can choose between [Tailscale](https://tailscale.com),
[Wireguard](https://www.wireguard.com), or use both like me. I primarily use
WireGuard but keep Tailscale as a backup for when the WireGuard server is down.

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

You may want to back up the `external/terraform.tfvars` file to a secure location.

Apply the secret:

```sh
make external
```

Finally, [enable subnet routes](https://tailscale.com/kb/1019/subnets#step-3-enable-subnet-routes-from-the-admin-console) for `homelab-router`
from the [admin console](https://login.tailscale.com/admin/machines).

You can now connect to your homelab via Tailscale and [invite user to your Tailscale network](https://tailscale.com/kb/1371/invite-users).

## Wireguard (requires port-forwarding)

### Prerequisites

Find your public IP address using:

```sh
curl -4 ifconfig.me
```

If you don’t have a static IP address, use dynamic DNS and replace the IP with
your domain name.

Next, configure port forwarding in your router for the WireGuard service.

!!! example

    Each router is different, here's mine for reference:

    - Protocol: `UDP`
    - Start Port: `51820`
    - End Port: `51820`
    - Local IP Address: `192.168.1.226` (find it with `kubectl get service -n wireguard wireguard`)
    - Start Port Local: `51820`
    - End Port Local: `51820`

Generate a key pair for the server:

```sh
wg genkey | tee /dev/tty | wg pubkey
```

This will generate a private key and a public key, in that order. Add the
private key to `external/terraform.tfvars` as an extra secret:

```hcl
extra_secrets = {
  wireguard-private-key = "privatekeyhere"
}
```

You may want to back up the `external/terraform.tfvars` file to a secure location.

Apply the secret:

```sh
make external
```

I use `172.16.0.0/12` as the private IP range for WireGuard, but you can choose
any private IP address range you prefer in `./apps/wireguard/values.yaml`. I
also recommend removing my peers and adding your own.

### Add a new device to the server

!!! info

    Each device requires its own configuration.

Generate a new key pair for the device. You can generate it for the user, or
they can generate it themselves if they prefer to keep the private key
confidential:

```sh
wg genkey | tee /dev/tty | wg pubkey
```

This will generate a private key and a public key, in that order. The private
key must be saved in a secure password manager, and save the public key for the
next step.

Update the list of peers in `./apps/wireguard/values.yaml`, make sure you
replace all of my peers with yours.

!!! example

    Example configuration for my phone:

    ```ini
    [Peer]
    PublicKey = nITHFdgTkNZOTWeSWqnGXjgwlCJMKRCnnUsjMx2yp2U=
    AllowedIPs = 172.16.0.12/32
    ```

    - The public key is the one generated in the previous step.
    - `172.16.0.12/32` is the device's private IP address, manually selected from
      the `172.16.0.0/12` range mentioned above.

### Add the Wireguard config to the device

Create a new configuration file for the device:

```ini
[Interface]
Address = <CLIENT PRIVATE IP>/32
PrivateKey = <CLIENT PRIVATE KEY>

[Peer]
PublicKey = <SERVER PUBLIC KEY>
Endpoint = <SERVER PUBLIC IP>:51820
AllowedIPs = <SERVER PRIVATE IP>/32, <LOAD BALANCER IP RANGE>
```

Replace placeholders with actual values and save as `wg0.conf`.

!!! example

    Example configuration for my phone:

    ```ini
    [Interface]
    Address = 172.16.0.12/32
    PrivateKey = <REDACTED>

    [Peer]
    PublicKey = sSAZS1Z3vB7Wx8e2yVqXfeHjgWTa80wnSYoma3mZkiU
    Endpoint = <HOME IP>:51820
    AllowedIPs = 172.16.0.1/32, 192.168.1.224/27
    ```

The client can now import this configuration and connect to your WireGuard
mesh. Make sure you clean up the `wg0.conf` file after importing it to the
client.

=== "Mobile"

    Generate a QR code from the configuration file:

    ```sh
    qrencode -t ansiutf8 -r wg0.conf
    ```

    Then scan the QR code using the official WireGuard app.

=== "Linux"

    Import the WireGuard configuration using NetworkManager:

    ```sh
    nmcli connection import type wireguard file wg0.conf
    ```

    Activate the connection:

    ```sh
    nmcli connection up wg0
    ```
