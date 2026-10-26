resource "kubectl_manifest" "cloudflare_ddns" {
  depends_on        = [kubectl_manifest.openbao]
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "cloudflare-ddns"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
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
                labels      = { "homelab.khuedoan.com/bao-secret" = "true" }
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
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
