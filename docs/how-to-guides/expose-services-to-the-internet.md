# Expose services to the internet

> [!NOTE]
>
> This tutorial is for Cloudflare Tunnel users, please skip if you use port forwarding.

Apply the environment's Cloudflare unit to create a tunnel if you haven't already,
then attach your `HTTPRoute` to the `public` Gateway (replace `example.com` with your domain):

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: website
  namespace: website
  annotations:
    external-dns.kubernetes.io/cloudflare-proxied: "true"
spec:
  parentRefs:
    - name: public
      namespace: istio-system
      sectionName: https
  hostnames:
    - www.example.com
  rules:
    - backendRefs:
        - name: website
          port: 3000
```

The `public` Gateway supplies the DNS target for the environment's tunnel.
