resource "kubectl_manifest" "renovate" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "renovate"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = var.repository.url
        targetRevision = var.repository.revision
        path           = "infra/_modules/platform/renovate"
        helm = {
          releaseName = "renovate"
          values = yamlencode({
            renovate = {
              # TODO: Revisit resource sizing after measuring usage.
              resources = { requests = { cpu = "50m", memory = "128Mi" } }
              cronjob = {
                schedule = "0 9 * * *"
              }
              existingSecret = "renovate-secret"
              renovate = {
                config = <<-EOT
{
  "platform": "gitea",
  "endpoint": "https://git.khuedoan.com/api/v1",
  "gitAuthor": "Renovate Bot <bot@renovateapp.com>",
  "autodiscover": true
}
EOT

              }
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "renovate" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
