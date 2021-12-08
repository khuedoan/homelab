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
  value   = "${cloudflare_argo_tunnel.homelab.id}.cfargotunnel.com"
  proxied = true
  ttl     = 1 # Auto
}

# TODO
# api token
# add it to certmanager, external-dns, cloudflaredknamespace
