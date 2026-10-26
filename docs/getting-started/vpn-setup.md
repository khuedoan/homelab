# VPN setup

You can choose between [Tailscale](https://tailscale.com),
[WireGuard](https://www.wireguard.com), or use both like me. I primarily use
WireGuard but keep Tailscale as a backup for when the WireGuard server is down.

## Tailscale (requires third-party account)

Tailscale provides access through a [subnet router](https://tailscale.com/docs/features/subnet-routers) in the cluster.


1. Select the environment's kubeconfig:

    ```sh
    export KUBECONFIG="$PWD/infra/production/kubeconfig.yaml"
    ```

2. Read the router's login URL:

    ```sh
    kubectl -n tailscale logs deployment/tailscale
    ```

3. Open the login URL and enroll the device in your tailnet.
4. Approve its advertised subnet in the Tailscale admin console.
5. Install Tailscale on your devices, sign in, and enable subnet routes where
   required by the client.


You can [invite users](https://tailscale.com/kb/1371/invite-users) to your tailnet.

## WireGuard (requires port-forwarding)

WireGuard Portal is available at `https://vpn.<domain>` with SSO.

### Prerequisites

Configure port forwarding in your router for the WireGuard service.

!!! example

    Each router is different, here's mine for reference:

    - Protocol: `UDP`
    - Start Port: `51820`
    - End Port: `51820`
    - Local IP Address: `192.168.1.226` (find it with `kubectl get service -n wireguard wireguard-wireguard`)
    - Start Port Local: `51820`
    - End Port Local: `51820`

### Configure the server

Sign in to the portal with an account in the `admins` group and create a server interface:

- Name: `wg0`
- Address: `10.44.0.1/24`
- Listen port: `51820`
- Enable default peer creation
- Client network: `10.44.0.0/24`
- Client endpoint: `wg.<domain>:51820`
- Client allowed IPs: `10.44.0.0/24` and your application LoadBalancer subnet

Choose a tunnel network that doesn't overlap your LAN, pod, or Service networks.

### Add a new device

WireGuard Portal will create one device by default. Users can add more if needed.

Install [WireGuard](https://www.wireguard.com/install/), then download the configuration as `wg0.conf` or scan its QR code.
Keep the configuration private.

=== "Mobile"

    Scan the QR code from the portal using the official WireGuard app.

=== "Linux"

    Import the WireGuard configuration using NetworkManager:

    ```sh
    nmcli connection import type wireguard file wg0.conf
    ```

    Activate the connection:

    ```sh
    nmcli connection up wg0
    ```

Connect and test access to a private application while outside your LAN.

## Headscale

[Headscale](https://headscale.net) is a self-hosted implementation of the
Tailscale control server that does not require a third-party account. To set it
up, take a look at the changes proposed by [this PR](https://github.com/khuedoan/homelab/pull/190).
