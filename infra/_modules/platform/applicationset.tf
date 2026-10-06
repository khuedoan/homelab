resource "kubectl_manifest" "apps" {
  server_side_apply = true

  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "ApplicationSet"
    metadata = {
      name      = "apps"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
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
          name   = "{{ .path.filename | trimSuffix \".yaml\" }}"
          labels = { "app.kubernetes.io/part-of" = "homelab" }
        }
        spec = {
          project = "default"
          sources = [{
            repoURL        = "https://bjw-s-labs.github.io/helm-charts"
            chart          = "app-template"
            targetRevision = "5.2.1"
            helm = {
              releaseName = "{{ .path.filename | trimSuffix \".yaml\" }}"
              valueFiles  = ["$values/apps/{{ .path.filename }}"]
            }
            }, {
            repoURL        = var.repository.url
            targetRevision = var.repository.revision
            ref            = "values"
          }]
          destination = {
            server    = "https://kubernetes.default.svc"
            namespace = "{{ .path.filename | trimSuffix \".yaml\" }}"
          }
          ignoreDifferences = local.ignore_differences
          syncPolicy        = local.sync_policy
        }
      }
    }
  })

  lifecycle {
    prevent_destroy = true
  }
}
