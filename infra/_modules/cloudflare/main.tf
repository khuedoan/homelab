data "cloudflare_zone" "zone" {
  filter = {
    name    = var.zone_name
    account = { id = var.cloudflare_account_id }
  }
}

data "cloudflare_account_api_token_permission_groups_list" "dns" {
  account_id = var.cloudflare_account_id
  scope      = "com.cloudflare.api.account.zone"
}

locals {
  dns_policies = [{
    effect = "allow"
    permission_groups = [
      for group in data.cloudflare_account_api_token_permission_groups_list.dns.result : { id = group.id }
      if contains(["Zone Read", "DNS Write"], group.name)
    ]
    resources = jsonencode({
      "com.cloudflare.api.account.zone.${data.cloudflare_zone.zone.id}" = "*"
    })
  }]
}

resource "random_password" "tunnel_secret" {
  length  = 64
  special = false
}

resource "cloudflare_zero_trust_tunnel_cloudflared" "homelab" {
  account_id    = var.cloudflare_account_id
  name          = "${var.resource_prefix}-tunnel"
  config_src    = "local"
  tunnel_secret = base64encode(random_password.tunnel_secret.result)
}

# Not proxied, not accessible. Just a record for auto-created CNAMEs by external-dns.
resource "cloudflare_dns_record" "tunnel" {
  zone_id = data.cloudflare_zone.zone.id
  type    = "CNAME"
  name    = "${var.resource_prefix}-tunnel.${var.domain}"
  content = "${cloudflare_zero_trust_tunnel_cloudflared.homelab.id}.cfargotunnel.com"
  proxied = false
  ttl     = 1 # Auto
}

resource "cloudflare_account_token" "external_dns" {
  account_id = var.cloudflare_account_id
  name       = "${var.resource_prefix}-external-dns"
  policies   = local.dns_policies
}

resource "cloudflare_account_token" "cert_manager" {
  account_id = var.cloudflare_account_id
  name       = "${var.resource_prefix}-cert-manager"
  policies   = local.dns_policies
}
