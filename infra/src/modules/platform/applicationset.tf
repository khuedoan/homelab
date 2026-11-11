resource "kubernetes_manifest" "apps" {
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "ApplicationSet"
    metadata = {
      name      = "apps"
      namespace = "argocd"
    }
    spec = {
      goTemplate        = true
      goTemplateOptions = ["missingkey=error"]
      generators = [{
        git = {
          repoURL  = var.repository.url
          revision = var.repository.revision
          files    = [{ path = "apps/*.yaml" }]
        }
      }]
      template = {
        metadata = {
          name        = "{{ .path.filename | trimSuffix \".yaml\" }}"
          annotations = { "argocd.argoproj.io/compare-options" = "IncludeMutationWebhook=true" }
        }
        spec = {
          project = kubernetes_manifest.apps_project.manifest.metadata.name
          sources = [{
            repoURL        = "ghcr.io/bjw-s-labs/helm"
            chart          = "app-template"
            targetRevision = "5.3.0"
            helm = {
              releaseName = "{{ .path.filename | trimSuffix \".yaml\" }}"
              valueFiles  = ["$values/apps/{{ .path.filename }}"]
              parameters = [{
                name  = "global.domain"
                value = var.domain
              }]
            }
            }, {
            repoURL        = var.repository.url
            targetRevision = var.repository.revision
            ref            = "values"
          }]
          destination = {
            name      = "in-cluster"
            namespace = "{{ .path.filename | trimSuffix \".yaml\" }}"
          }
          ignoreDifferences = local.ignore_differences
          syncPolicy = merge(local.sync_policy, {
            automated = {
              enabled  = !var.recovery_mode
              prune    = true
              selfHeal = true
            }
            managedNamespaceMetadata = {
              labels = { "istio.io/dataplane-mode" = "ambient" }
            }
          })
        }
      }
    }
  }

  lifecycle {
    prevent_destroy = true
  }
}
