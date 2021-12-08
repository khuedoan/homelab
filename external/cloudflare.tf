provider "cloudflare" {
  # Environment variables
  # CLOUDFLARE_API_TOKEN
}

data "cloudflare_zone" "khuedoan_com" {
  name = "khuedoan.com"
}

resource "random_password" "tunnel_secret" {
  length  = 64
  special = false
}

resource "cloudflare_argo_tunnel" "homelab" {
  # TODO (optimize) Use variable for account_id
  account_id = "xxx"
  name       = "homelab"
  secret     = base64encode(random_password.tunnel_secret.result)
}

resource "cloudflare_record" "tunnels" {
  for_each = toset([
    "git"
  ])

  zone_id = data.cloudflare_zone.khuedoan_com.id
  type    = "CNAME"
  name    = each.key
  value   = cloudflare_argo_tunnel.homelab.cname
  proxied = true
  ttl     = 1 # Auto
}

# TODO
# api token
# add it to certmanager, external-dns, cloudflaredknamespace

resource "kubernetes_namespace" "namespaces" {
  for_each = toset([
    "cert-manager",
    "cloudflared",
    "external-dns",
    "velero"
  ])

  metadata {
    name = each.key
  }
}

resource "kubernetes_secret" "cloudflared_credentials" {
  metadata {
    name = "cloudflared-credentials"
    namespace = "cloudflared"
  }

  data = {
    "credentials.json" = base64encode(jsonencode({
      AccountTag   = "" # TODO account_id
      TunnelName   = cloudflare_argo_tunnel.homelab.name
      TunnelID     = cloudflare_argo_tunnel.homelab.id
      TunnelSecret = base64encode(random_password.tunnel_secret.result)
    }))
  }
}
