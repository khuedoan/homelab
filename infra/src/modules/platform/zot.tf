resource "kubernetes_manifest" "zot" {
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "zot"
      namespace = "argocd"
    }
    spec = {
      project = kubernetes_manifest.system_project.manifest.metadata.name
      sources = [{
        repoURL        = "https://zotregistry.dev/helm-charts"
        chart          = "zot"
        targetRevision = "0.1.128"
        helm = {
          releaseName = "zot"
          values = yamlencode({
            # TODO: Revisit resource sizing after measuring usage.
            resources       = { requests = { cpu = "50m", memory = "128Mi" } }
            persistence     = true
            serviceHeadless = { enabled = true }
            pvc = {
              create  = true
              storage = "2Gi"
            }
            mountConfig = true
            configFiles = { "config.json" = jsonencode({
              storage = { rootDirectory = "/var/lib/registry" }
              http = {
                address     = "0.0.0.0"
                port        = "5000"
                externalUrl = "https://registry.${var.domain}"
                auth = {
                  openid = {
                    providers = {
                      oidc = {
                        name     = "SSO"
                        issuer   = "https://auth.${var.domain}/oauth2/openid/zot"
                        clientid = "zot"
                        scopes   = ["openid", "profile", "email", "groups"]
                      }
                    }
                  }
                }
                accessControl = {
                  repositories = {
                    "**" = {
                      anonymousPolicy = ["read"]
                      defaultPolicy   = ["read", "create", "update", "delete"]
                    }
                  }
                }
              }
              extensions = { search = { enable = true }, ui = { enable = true } }
              log        = { level = "info" }
            }) }
            httproute = {
              enabled = true
              parentRefs = [{
                name        = "gateway"
                namespace   = "istio-system"
                sectionName = "https"
              }]
              hostnames = ["registry.${var.domain}"]
              pathType  = "PathPrefix"
              path      = "/"
            }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "zot" }
      ignoreDifferences = local.ignore_differences
      syncPolicy = merge(local.sync_policy, {
        automated = {
          enabled  = !var.recovery_mode
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

resource "kubernetes_manifest" "zot_resources" {
  depends_on = [kubernetes_manifest.openbao]
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name        = "zot-resources"
      namespace   = "argocd"
      annotations = { "argocd.argoproj.io/compare-options" = "IncludeMutationWebhook=true" }
    }
    spec = {
      project = kubernetes_manifest.system_project.manifest.metadata.name
      sources = [{
        repoURL        = "https://bjw-s-labs.github.io/helm-charts"
        chart          = "app-template"
        targetRevision = "5.2.1"
        helm = {
          releaseName = "zot-resources"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            secrets = {
              admin-secret = {
                forceRename = "registry-admin-secret"
                stringData = {
                  username = "admin"
                  password = "bao:secret/data/zot/admin#password"
                }
              }
            }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "zot" }
      ignoreDifferences = local.ignore_differences
      syncPolicy = merge(local.sync_policy, {
        automated = {
          enabled  = !var.recovery_mode
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
