# DNS setup

Because everyone DNS setup are different, DNS automation is not in the scope of the project.

Some options for DNS config (choose one):

- Update the `/etc/hosts` file (suitable for a test environment)
- Change the DNS config in your router
- Change the DNS config at your domain registrar (doesn't work with the [`home.arpa`](https://datatracker.ietf.org/doc/html/rfc8375) domain)

Before continuing to the next section for some examples, run this command to get a list of subdomain and its address:

```sh
./scripts/get-dns-config
```

## Update `/etc/hosts`

You paste the output from the command above directly to your `/etc/hosts` file like this:

```
# Static table lookup for hostnames.
# See hosts(5) for details.

192.168.1.150 argocd.mydomain.com
192.168.1.150 git.mydomain.com
192.168.1.150 jellyfin.mydomain.com
# etc.
```

## In your router

You can add each subdomain one by one like the previous method, or use a wildcard `*.mydomain.com` and point it to the IP address of the load blancer (from the output of the previous command)

## At your domain registrar

For example here is my set up for Cloudflare, automated using Terraform:

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
  value   = "192.168.1.150" # NGINX LoadBalancer IP
  ttl     = 1 # Auto
}
```
