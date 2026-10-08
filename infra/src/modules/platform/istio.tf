resource "kubectl_manifest" "istio" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "istio"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://istio-release.storage.googleapis.com/charts"
        targetRevision = "1.30.5"
        chart          = "base"
        helm           = { releaseName = "istio" }
        }, {
        repoURL        = "https://istio-release.storage.googleapis.com/charts"
        targetRevision = "1.30.5"
        chart          = "istiod"
        helm = {
          releaseName = "istio"
          values = yamlencode({
            # TODO: Revisit resource sizing after measuring usage.
            resources = { requests = { cpu = "100m", memory = "128Mi" } }
          })
        }
      }]
      destination = { server = "https://kubernetes.default.svc", namespace = "istio-system" }
      ignoreDifferences = concat(local.ignore_differences, [
        for name in ["istio-validator-istio-system", "istiod-default-validator"] : {
          group             = "admissionregistration.k8s.io"
          kind              = "ValidatingWebhookConfiguration"
          name              = name
          jqPathExpressions = [".webhooks[].failurePolicy", ".webhooks[].clientConfig.caBundle"]
        }
      ])
      syncPolicy = merge(local.sync_policy, {
        syncOptions = concat(local.sync_policy.syncOptions, ["RespectIgnoreDifferences=true"])
      })
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}

resource "kubectl_manifest" "istio_resources" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "istio-resources"
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
          releaseName = "istio-resources"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            configMaps = {
              public-gateway = {
                forceRename = "public-gateway"
                data        = { service = yamlencode({ spec = { type = "ClusterIP" } }) }
              }
            }
            rawResources = {
              gateway = {
                forceRename = "gateway"
                manifest = {
                  apiVersion = "gateway.networking.k8s.io/v1"
                  kind       = "Gateway"
                  metadata = {
                    namespace   = "istio-system"
                    annotations = { "cert-manager.io/cluster-issuer" = "letsencrypt-prod" }
                  }
                  spec = {
                    gatewayClassName = "istio"
                    listeners = [{
                      name          = "http"
                      port          = 80
                      protocol      = "HTTP"
                      allowedRoutes = { namespaces = { from = "All" } }
                      }, {
                      name          = "https"
                      hostname      = "*.khuedoan.com"
                      port          = 443
                      protocol      = "HTTPS"
                      tls           = { mode = "Terminate", certificateRefs = [{ name = "wildcard-tls" }] }
                      allowedRoutes = { namespaces = { from = "All" } }
                    }]
                  }
                }
              }
              public = {
                forceRename = "public"
                manifest = {
                  apiVersion = "gateway.networking.k8s.io/v1"
                  kind       = "Gateway"
                  metadata = {
                    namespace   = "istio-system"
                    annotations = { "external-dns.alpha.kubernetes.io/target" = "homelab-tunnel.khuedoan.com" }
                  }
                  spec = {
                    gatewayClassName = "istio"
                    infrastructure   = { parametersRef = { group = "", kind = "ConfigMap", name = "public-gateway" } }
                    listeners = [{
                      name          = "https"
                      hostname      = "*.khuedoan.com"
                      port          = 443
                      protocol      = "HTTPS"
                      tls           = { mode = "Terminate", certificateRefs = [{ name = "wildcard-tls" }] }
                      allowedRoutes = { namespaces = { from = "All" } }
                    }]
                  }
                }
              }
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "istio-system" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
