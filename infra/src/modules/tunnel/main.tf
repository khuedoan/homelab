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
  zone_id = var.zone_id
  type    = "CNAME"
  name    = "${var.resource_prefix}-tunnel.${var.domain}"
  content = "${cloudflare_zero_trust_tunnel_cloudflared.homelab.id}.cfargotunnel.com"
  proxied = false
  ttl     = 1 # Auto
}
