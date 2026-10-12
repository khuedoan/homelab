resource "kubectl_manifest" "zot" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "zot"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://zotregistry.dev/helm-charts"
        targetRevision = "0.1.128"
        chart          = "zot"
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
      destination       = { server = "https://kubernetes.default.svc", namespace = "zot" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}

resource "kubectl_manifest" "zot_resources" {
  depends_on        = [kubectl_manifest.openbao]
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "zot-resources"
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
          releaseName = "zot-resources"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            secrets = {
              admin-secret = {
                forceRename = "registry-admin-secret"
                labels      = { "homelab.khuedoan.com/bao-secret" = "true" }
                stringData = {
                  username = "admin"
                  password = "bao:secret/data/registry.admin#password"
                }
              }
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "zot" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
