resource "kubectl_manifest" "external_secrets" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "external-secrets"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://charts.external-secrets.io"
        chart          = "external-secrets"
        targetRevision = "0.10.2"
        helm = {
          releaseName = "external-secrets"
          values = yamlencode({
            # TODO: Revisit resource sizing after measuring usage.
            resources      = { requests = { cpu = "10m", memory = "64Mi" } }
            webhook        = { resources = { requests = { cpu = "10m", memory = "32Mi" } } }
            certController = { resources = { requests = { cpu = "10m", memory = "32Mi" } } }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "external-secrets" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
