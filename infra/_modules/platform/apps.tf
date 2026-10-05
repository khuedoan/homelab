resource "kubectl_manifest" "apps" {
  for_each = var.apps

  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = each.key
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = var.repository.url
        targetRevision = var.repository.revision
        path           = each.value.path
        helm = {
          releaseName = each.key
          values      = each.value.values
        }
      }]
      destination = {
        server    = "https://kubernetes.default.svc"
        namespace = each.value.namespace
      }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })

  lifecycle {
    prevent_destroy = true
  }
}
