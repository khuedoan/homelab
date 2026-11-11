resource "kubernetes_manifest" "external_dns" {
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "external-dns"
      namespace = "argocd"
    }
    spec = {
      project = kubernetes_manifest.system_project.manifest.metadata.name
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
            extraArgs          = ["--annotation-filter=external-dns.kubernetes.io/exclude notin (true)", "--zone-id-filter=$(CF_ZONE_ID)"]
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
      destination       = { name = "in-cluster", namespace = "external-dns" }
      ignoreDifferences = local.ignore_differences
      syncPolicy = merge(local.sync_policy, {
        automated = {
          enabled  = !var.recovery_mode
          prune    = true
          selfHeal = true
        }
      })
    }
  }
  lifecycle {
    prevent_destroy = true
  }
}

resource "kubernetes_manifest" "external_dns_resources" {
  depends_on = [kubernetes_manifest.openbao]
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name        = "external-dns-resources"
      namespace   = "argocd"
      annotations = { "argocd.argoproj.io/compare-options" = "IncludeMutationWebhook=true" }
    }
    spec = {
      project = kubernetes_manifest.system_project.manifest.metadata.name
      sources = [{
        repoURL        = "ghcr.io/bjw-s-labs/helm"
        chart          = "app-template"
        targetRevision = "5.3.0"
        helm = {
          releaseName = "external-dns-resources"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            secrets = {
              cloudflare = {
                forceRename = "cloudflare-api-token"
                stringData = {
                  "api-token" = "bao:secret/data/external-dns/cloudflare#api_token"
                  "zone-id"   = "bao:secret/data/external-dns/cloudflare#zone_id"
                }
              }
            }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "external-dns" }
      ignoreDifferences = local.ignore_differences
      syncPolicy = merge(local.sync_policy, {
        automated = {
          enabled  = !var.recovery_mode
          prune    = true
          selfHeal = true
        }
      })
    }
  }
  lifecycle {
    prevent_destroy = true
  }
}
