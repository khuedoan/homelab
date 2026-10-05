resource "kubectl_manifest" "global_secrets" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "global-secrets"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = var.repository.url
        targetRevision = var.repository.revision
        path           = "infra/_modules/platform/global-secrets"
        helm = {
          releaseName = "global-secrets"
          values = yamlencode({
            secretDefinitions = [{
              data = [{
                key     = "password"
                length  = 32
                special = true
              }]
              name = "forgejo.admin"
              }, {
              data = [{
                key     = "client_secret"
                length  = 32
                special = false
              }]
              name = "dex.grafana"
              }, {
              data = [{
                key     = "client_secret"
                length  = 32
                special = false
              }]
              name = "dex.forgejo"
              }, {
              data = [{
                key     = "password"
                length  = 32
                special = true
              }]
              name = "registry.admin"
              }, {
              data = [{
                key     = "secret"
                length  = 32
                special = false
              }]
              name = "woodpecker.agent"
              }, {
              data = [{
                key     = "PAPERLESS_ADMIN_PASSWORD"
                length  = 32
                special = true
              }]
              name = "paperless.admin"
            }]
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "global-secrets" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
