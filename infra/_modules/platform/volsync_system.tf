resource "kubectl_manifest" "volsync_system" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "volsync-system"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://backube.github.io/helm-charts"
        chart          = "volsync"
        targetRevision = "0.9.1"
        helm = {
          releaseName = "volsync-system"
          values = yamlencode({
            # TODO: Revisit resource sizing after measuring usage.
            resources = { requests = { cpu = "10m", memory = "64Mi" } }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "volsync-system" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
