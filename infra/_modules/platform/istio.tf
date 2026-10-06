resource "kubectl_manifest" "istio" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "istio"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = var.repository.url
        targetRevision = var.repository.revision
        path           = "infra/_modules/platform/istio"
        helm           = { releaseName = "istio" }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "istio-system" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
