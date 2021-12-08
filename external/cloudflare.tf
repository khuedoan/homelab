locals {
  internal_records = [
    "*.knative",
    "argocd",
    "authentik",
    "dex",
    "grafana",
    "home",
    "jellyfin",
    "paperless",
    "seafile",
    "syncthing",
    "tekton",
    "vault",
  ]

  tunnel_records = [
    "git"
  ]
}

resource "cloudflare_record" "internal_records" {
  for_each = toset(local.internal_records)
  zone_id = cloudflare_zone.khuedoan_com.id
  type    = "A"
  name    = each.key
  # TODO use data to get ingress IP
  value   = "192.168.1.150"
  ttl     = 1 # Auto
}

resource "random_password" "homelab_tunnel" {
  length  = 64
  special = false
}

resource "cloudflare_argo_tunnel" "homelab" {
  # TODO (optimize) Use variable for account_id
  account_id = "xxx"
  name       = "homelab"
  secret     = base64encode(random_password.homelab_tunnel.result)
}

resource "cloudflare_record" "git" {
  zone_id = cloudflare_zone.khuedoan_com.id
  type    = "CNAME"
  name    = "git"
  value   = "${cloudflare_argo_tunnel.homelab.id}.cfargotunnel.com"
  proxied = true
  ttl     = 1 # Auto
}
