output "secrets" {
  sensitive = true
  value = {
    "cert-manager/cloudflare" = {
      api_token = cloudflare_account_token.cert_manager.value
    }
    "external-dns/cloudflare" = {
      api_token = cloudflare_account_token.external_dns.value
      zone_id   = data.cloudflare_zone.zone.id
    }
    "cloudflare-ddns/cloudflare" = {
      api_token = cloudflare_account_token.ddns.value
    }
    "cloudflared/cloudflare" = {
      "credentials.json" = jsonencode({
        AccountTag   = var.cloudflare_account_id
        TunnelName   = cloudflare_zero_trust_tunnel_cloudflared.homelab.name
        TunnelID     = cloudflare_zero_trust_tunnel_cloudflared.homelab.id
        TunnelSecret = base64encode(random_password.tunnel_secret.result)
      })
    }
  }
}
