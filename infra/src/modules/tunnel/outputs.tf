output "secrets" {
  sensitive = true
  value = {
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
