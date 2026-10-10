output "tunnel_credentials" {
  sensitive = true
  value = jsonencode({
    AccountTag   = var.cloudflare_account_id
    TunnelName   = cloudflare_zero_trust_tunnel_cloudflared.homelab.name
    TunnelID     = cloudflare_zero_trust_tunnel_cloudflared.homelab.id
    TunnelSecret = base64encode(random_password.tunnel_secret.result)
  })
}

output "external_dns_token" {
  value     = cloudflare_account_token.external_dns.value
  sensitive = true
}

output "cert_manager_token" {
  value     = cloudflare_account_token.cert_manager.value
  sensitive = true
}

output "zone_id" {
  value = data.cloudflare_zone.zone.id
}
