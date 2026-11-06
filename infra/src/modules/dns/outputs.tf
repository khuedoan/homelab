output "zone_id" {
  value = data.cloudflare_zone.zone.id
}

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
  }
}
