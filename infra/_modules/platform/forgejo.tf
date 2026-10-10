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
        repoURL        = "code.forgejo.org/forgejo-helm"
        targetRevision = "17.1.7"
        chart          = "forgejo"
        helm = {
          releaseName = "forgejo"
          values = yamlencode({
            fullnameOverride = "forgejo"
            image            = { tag = "16.0.5" }
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
                  DOMAIN       = "git.${var.domain}"
                  ROOT_URL     = "https://git.${var.domain}"
                  LANDING_PAGE = "explore"
                  OFFLINE_MODE = true
                }
                webhook = { ALLOWED_HOST_LIST = "private" }
              }
            }
            persistence = {
              claimName = "forgejo-data"
              size      = "2Gi"
            }
            httpRoute = {
              enabled = true
              parentRefs = [{
                name        = "gateway"
                namespace   = "istio-system"
                sectionName = "https"
              }]
              hostnames = ["git.${var.domain}"]
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

resource "kubectl_manifest" "forgejo_resources" {
  depends_on        = [kubectl_manifest.openbao]
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "forgejo-resources"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://bjw-s-labs.github.io/helm-charts"
        targetRevision = "5.2.1"
        chart          = "app-template"
        helm = {
          releaseName = "forgejo-resources"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            rawResources = {
              admin = {
                forceRename = "forgejo-admin"
                manifest = {
                  apiVersion = "v1"
                  kind       = "Secret"
                  metadata = {
                    labels = { "homelab.khuedoan.com/bao-secret" = "true" }
                    annotations = {
                      "argocd.argoproj.io/sync-wave"                               = "-1"
                      "secrets-webhook.security.bank-vaults.io/provider"           = "bao"
                      "secrets-webhook.security.bank-vaults.io/bao-role"           = "forgejo"
                      "secrets-webhook.security.bank-vaults.io/bao-path"           = "kubernetes"
                      "secrets-webhook.security.bank-vaults.io/bao-serviceaccount" = "default"
                    }
                  }
                  data = {
                    username = base64encode("forgejo_admin")
                    password = base64encode("bao:secret/data/forgejo.admin#password")
                  }
                }
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
