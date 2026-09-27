output "tunnel_credentials" {
  sensitive = true
  value = jsonencode({
    AccountTag   = var.cloudflare_account_id
    TunnelName   = cloudflare_tunnel.homelab.name
    TunnelID     = cloudflare_tunnel.homelab.id
    TunnelSecret = base64encode(random_password.tunnel_secret.result)
  })
}

output "external_dns_token" {
  value     = cloudflare_api_token.external_dns.value
  sensitive = true
}

output "cert_manager_token" {
  value     = cloudflare_api_token.cert_manager.value
  sensitive = true
}
