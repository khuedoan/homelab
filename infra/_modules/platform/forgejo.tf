resource "kubectl_manifest" "forgejo" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "forgejo"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = var.repository.url
        targetRevision = var.repository.revision
        path           = "infra/_modules/platform/forgejo"
        helm = {
          releaseName = "forgejo"
          values = yamlencode({
            configuration = {
              organizations = [{
                name        = "ops"
                description = "Operations"
              }]
              repositories = [
                {
                  name    = "homelab"
                  owner   = "ops"
                  private = false
                  migrate = { source = "https://github.com/khuedoan/homelab", mirror = false }
                },
                {
                  name    = "blog"
                  owner   = "khuedoan"
                  migrate = { source = "https://github.com/khuedoan/blog", mirror = true }
                },
                {
                  name    = "backstage"
                  owner   = "khuedoan"
                  migrate = { source = "https://github.com/khuedoan/backstage", mirror = true }
                },
              ]
            }
            forgejo = {
              image = { tag = "16.0.5" }
              # TODO: Revisit resource sizing after measuring usage.
              resources = { requests = { cpu = "50m", memory = "128Mi" } }
              gitea = {
                admin = { existingSecret = "forgejo-admin" }
                config = {
                  database = { DB_TYPE = "sqlite3" }
                  actions  = { ENABLED = false }
                  repository = {
                    DEFAULT_BRANCH      = "master"
                    DISABLED_REPO_UNITS = "repo.wiki,repo.projects,repo.packages"
                    DISABLE_STARS       = true
                  }
                  "service.explore" = { DISABLE_USERS_PAGE = true }
                  server = {
                    DOMAIN       = "git.khuedoan.com"
                    ROOT_URL     = "https://git.khuedoan.com"
                    LANDING_PAGE = "explore"
                    OFFLINE_MODE = true
                  }
                  webhook = { ALLOWED_HOST_LIST = "private" }
                }
              }
              persistence = {
                claimName = "forgejo-data"
                size      = "2Gi"
                # TODO: Replace temporary local-path storage with replicated storage.
                storageClass = "local-path"
              }
              ingress = {
                enabled   = true
                className = "nginx"
                annotations = {
                  "cert-manager.io/cluster-issuer" = "letsencrypt-prod"
                }
                hosts = [{
                  host  = "git.khuedoan.com"
                  paths = [{ path = "/", pathType = "Prefix" }]
                }]
                tls = [{
                  hosts      = ["git.khuedoan.com"]
                  secretName = "forgejo-tls"
                }]
              }
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "forgejo" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
