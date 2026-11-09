resource "kubernetes_manifest" "loki" {
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "loki"
      namespace = "argocd"
    }
    spec = {
      project = kubernetes_manifest.system_project.manifest.metadata.name
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
              }
            }
            promtail = {
              resources = { requests = { cpu = "10m", memory = "32Mi" } }
            }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "loki" }
      ignoreDifferences = local.ignore_differences
      syncPolicy = merge(local.sync_policy, {
        automated = {
          enabled  = true
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
