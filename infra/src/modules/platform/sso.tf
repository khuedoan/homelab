resource "kubernetes_manifest" "argocd_resources" {
  depends_on = [kubernetes_manifest.openbao]
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name        = "argocd-resources"
      namespace   = "argocd"
      annotations = { "argocd.argoproj.io/compare-options" = "IncludeMutationWebhook=true" }
    }
    spec = {
      project = kubernetes_manifest.system_project.manifest.metadata.name
      sources = [{
        repoURL        = "https://bjw-s-labs.github.io/helm-charts"
        targetRevision = "5.2.1"
        chart          = "app-template"
        helm = {
          releaseName = "argocd-resources"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            secrets = {
              sso = {
                forceRename = "argocd-sso"
                labels = {
                  "app.kubernetes.io/part-of" = "argocd"
                }
                stringData = { client_secret = "bao:secret/data/argocd/sso#client_secret" }
              }
            }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "argocd" }
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
