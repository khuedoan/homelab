resource "kubectl_manifest" "loki" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "loki"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://grafana.github.io/helm-charts"
        chart          = "loki-stack"
        targetRevision = "2.10.1"
        helm = {
          releaseName = "loki"
          values = yamlencode({
            loki = {
              # TODO: Revisit resource sizing after measuring usage.
              resources = { requests = { cpu = "50m", memory = "128Mi" } }
              serviceMonitor = {
                enabled = true
              }
              persistence = {
                enabled = true
                size    = "2Gi"
                # TODO: Replace temporary local-path storage with replicated storage.
                storageClassName = "local-path"
              }
            }
            promtail = { resources = { requests = { cpu = "10m", memory = "32Mi" } } }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "loki" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
