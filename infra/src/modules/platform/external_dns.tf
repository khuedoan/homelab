resource "kubectl_manifest" "external_dns" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "external-dns"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://kubernetes-sigs.github.io/external-dns"
        chart          = "external-dns"
        targetRevision = "1.23.0"
        helm = {
          releaseName = "external-dns"
          values = yamlencode({
            # TODO: Revisit resource sizing after measuring usage.
            resources = { requests = { cpu = "10m", memory = "32Mi" } }
            env = [for key, name in { "api-token" = "CF_API_TOKEN", "zone-id" = "CF_ZONE_ID" } : {
              name = name
              valueFrom = {
                secretKeyRef = {
                  key  = key
                  name = "cloudflare-api-token"
                }
              }
            }]
            extraArgs          = ["--annotation-filter=external-dns.alpha.kubernetes.io/exclude notin (true)", "--zone-id-filter=$(CF_ZONE_ID)"]
            domainFilters      = [var.domain]
            interval           = "5m"
            policy             = "upsert-only"
            serviceMonitor     = { enabled = true }
            provider           = { name = "cloudflare" }
            sources            = ["service", "gateway-httproute"]
            triggerLoopOnEvent = true
            txtOwnerId         = var.resource_prefix
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "external-dns" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}

resource "kubectl_manifest" "external_dns_resources" {
  depends_on        = [kubectl_manifest.openbao]
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "external-dns-resources"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://bjw-s-labs.github.io/helm-charts"
        targetRevision = "5.2.1"
        chart          = "app-template"
        helm = {
          releaseName = "external-dns-resources"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            secrets = {
              cloudflare = {
                forceRename = "cloudflare-api-token"
                labels      = { "homelab.khuedoan.com/bao-secret" = "true" }
                annotations = {
                  "secrets-webhook.security.bank-vaults.io/provider"           = "bao"
                  "secrets-webhook.security.bank-vaults.io/bao-role"           = "external-dns"
                  "secrets-webhook.security.bank-vaults.io/bao-path"           = "kubernetes"
                  "secrets-webhook.security.bank-vaults.io/bao-serviceaccount" = "default"
                }
                stringData = {
                  "api-token" = "bao:secret/data/infra/cloudflare/external_dns_token#value"
                  "zone-id"   = "bao:secret/data/infra/cloudflare/zone_id#value"
                }
              }
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "external-dns" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
