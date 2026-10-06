resource "kubectl_manifest" "woodpecker" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "woodpecker"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = var.repository.url
        targetRevision = var.repository.revision
        path           = "infra/_modules/platform/woodpecker"
        helm = {
          releaseName = "woodpecker"
          values = yamlencode({
            woodpecker = {
              agent = {
                # TODO: Revisit resource sizing after measuring usage.
                resources = { requests = { cpu = "10m", memory = "64Mi" } }
                env = {
                  WOODPECKER_BACKEND_K8S_STORAGE_RWX = false
                  WOODPECKER_MAX_WORKFLOWS           = 10
                }
                persistence = {
                  # TODO: Replace temporary local-path storage with replicated storage.
                  storageClass = "local-path"
                }
                replicaCount = 2
              }
              server = {
                resources = { requests = { cpu = "50m", memory = "128Mi" } }
                persistentVolume = {
                  size = "2Gi"
                  # TODO: Replace temporary local-path storage with replicated storage.
                  storageClass = "local-path"
                }
                env = {
                  WOODPECKER_ADMIN               = "forgejo_admin"
                  WOODPECKER_GITEA               = true
                  WOODPECKER_GITEA_URL           = "https://git.khuedoan.com"
                  WOODPECKER_HOST                = "https://ci.khuedoan.com"
                  WOODPECKER_OPEN                = true
                  WOODPECKER_EXPERT_WEBHOOK_HOST = "http://woodpecker-server.woodpecker"
                }
              }
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "woodpecker" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
