resource "kubernetes_manifest" "cloudflare_ddns" {
  depends_on = [kubernetes_manifest.openbao]
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name        = "cloudflare-ddns"
      namespace   = "argocd"
      annotations = { "argocd.argoproj.io/compare-options" = "IncludeMutationWebhook=true" }
    }
    spec = {
      project = kubernetes_manifest.system_project.manifest.metadata.name
      sources = [{
        repoURL        = "https://bjw-s-labs.github.io/helm-charts"
        chart          = "app-template"
        targetRevision = "5.3.0"
        helm = {
          releaseName = "cloudflare-ddns"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            controllers = {
              main = {
                containers = {
                  main = {
                    image   = { repository = "ghcr.io/favonia/cloudflare-ddns", tag = "1.18.0" }
                    env     = { IP4_DOMAINS = "public.${var.domain}" }
                    envFrom = [{ secretRef = { name = "cloudflare-ddns-cloudflare" } }]
                    resources = {
                      requests = { cpu = "10m", memory = "32Mi" }
                      limits   = { memory = "128Mi" }
                    }
                  }
                }
              }
            }
            secrets = {
              cloudflare = {
                suffix      = "cloudflare"
                annotations = { "argocd.argoproj.io/sync-wave" = "-1" }
                stringData = {
                  CLOUDFLARE_API_TOKEN = "bao:secret/data/cloudflare-ddns/cloudflare#api_token"
                }
              }
            }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "cloudflare-ddns" }
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
