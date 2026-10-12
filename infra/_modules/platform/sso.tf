resource "kubectl_manifest" "argocd_resources" {
  depends_on        = [kubectl_manifest.openbao]
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "argocd-resources"
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
          releaseName = "argocd-resources"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            secrets = {
              sso = {
                forceRename = "argocd-sso"
                labels = {
                  "homelab.khuedoan.com/bao-secret" = "true"
                  "app.kubernetes.io/part-of"       = "argocd"
                }
                stringData = { client_secret = "bao:secret/data/sso/argocd#client_secret" }
              }
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "argocd" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
