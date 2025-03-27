data "cloudflare_zone" "zone" {
  name = "khuedoan.com"
}

data "cloudflare_api_token_permission_groups" "all" {}

resource "random_password" "tunnel_secret" {
  length  = 64
  special = false
}

resource "cloudflare_tunnel" "homelab" {
  account_id = var.cloudflare_account_id
  name       = "homelab"
  secret     = base64encode(random_password.tunnel_secret.result)
}

# Not proxied, not accessible. Just a record for auto-created CNAMEs by external-dns.
resource "cloudflare_record" "tunnel" {
  zone_id = data.cloudflare_zone.zone.id
  type    = "CNAME"
  name    = "homelab-tunnel"
  value   = "${cloudflare_tunnel.homelab.id}.cfargotunnel.com"
  proxied = false
  ttl     = 1 # Auto
}

resource "kubernetes_secret" "cloudflared_credentials" {
  metadata {
    name      = "cloudflared-credentials"
    namespace = "cloudflared"

    annotations = {
      "app.kubernetes.io/managed-by" = "Terraform"
    }
  }

  data = {
    "credentials.json" = jsonencode({
      AccountTag   = var.cloudflare_account_id
      TunnelName   = cloudflare_tunnel.homelab.name
      TunnelID     = cloudflare_tunnel.homelab.id
      TunnelSecret = base64encode(random_password.tunnel_secret.result)
    })
  }
}

resource "cloudflare_api_token" "external_dns" {
  name = "homelab_external_dns"

  policy {
    permission_groups = [
      data.cloudflare_api_token_permission_groups.all.zone["Zone Read"],
      data.cloudflare_api_token_permission_groups.all.zone["DNS Write"]
    ]
    resources = {
      "com.cloudflare.api.account.zone.*" = "*"
    }
  }
}

resource "kubernetes_secret" "external_dns_token" {
  metadata {
    name      = "cloudflare-api-token"
    namespace = "external-dns"

    annotations = {
      "app.kubernetes.io/managed-by" = "Terraform"
    }
  }

  data = {
    "value" = cloudflare_api_token.external_dns.value
  }
}

resource "cloudflare_api_token" "cert_manager" {
  name = "homelab_cert_manager"

  policy {
    permission_groups = [
      data.cloudflare_api_token_permission_groups.all.zone["Zone Read"],
      data.cloudflare_api_token_permission_groups.all.zone["DNS Write"]
    ]
    resources = {
      "com.cloudflare.api.account.zone.*" = "*"
    }
  }
}

resource "kubernetes_secret" "cert_manager_token" {
  metadata {
    name      = "cloudflare-api-token"
    namespace = "cert-manager"

    annotations = {
      "app.kubernetes.io/managed-by" = "Terraform"
    }
  }

  data = {
    "api-token" = cloudflare_api_token.cert_manager.value
  }
}
