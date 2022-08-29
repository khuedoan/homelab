# Expose services to the internet

!!! info

    This tutorial is for Cloudflare Tunnel users, please skip if you use port forwarding.

Apply the `./external` layer to create a tunnel if you haven't already,
then add the following annotations to your `Ingress` object (replace `example.com` with your domain):

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    external-dns.alpha.kubernetes.io/target: "homelab-tunnel.example.com"
    external-dns.alpha.kubernetes.io/cloudflare-proxied: "true"
# ...
```
