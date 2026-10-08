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
        targetRevision = "0.1.126"
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
              # TODO: Replace temporary local-path storage with replicated storage.
              storageClassName = "local-path"
            }
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
            rawResources = {
              admin-secret = {
                forceRename = "registry-admin-secret"
                manifest = {
                  apiVersion = "v1"
                  kind       = "Secret"
                  metadata = {
                    namespace = "zot"
                    labels    = { "homelab.khuedoan.com/bao-secret" = "true" }
                    annotations = {
                      "secrets-webhook.security.bank-vaults.io/provider"           = "bao"
                      "secrets-webhook.security.bank-vaults.io/bao-role"           = "zot"
                      "secrets-webhook.security.bank-vaults.io/bao-path"           = "kubernetes"
                      "secrets-webhook.security.bank-vaults.io/bao-serviceaccount" = "default"
                    }
                  }
                  data = {
                    username = base64encode("admin")
                    password = base64encode("bao:secret/data/registry.admin#password")
                  }
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
