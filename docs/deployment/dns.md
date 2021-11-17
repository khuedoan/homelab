# DNS setup

Some options for DNS config:

- Change the DNS config in your router
- Run your own DNS server in the cluster
- Change the DNS config in your domain registrar

For example here is my set up using the third option with Cloudflare, automated using Terraform:

```hcl
resource "cloudflare_record" "homelab_records" {
  for_each = toset([
    "*.knative",
    "argocd",
    "dex",
    "git",
    "grafana",
    "tekton",
    "vault",
    # etc.
  ])

  zone_id = cloudflare_zone.khuedoan_com.id
  type    = "A"
  name    = each.key
  value   = "192.168.1.150"
  ttl     = 1 # Auto
}
```
