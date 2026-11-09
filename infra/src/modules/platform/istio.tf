locals {
  istio_version = "1.30.5"
}

resource "kubernetes_manifest" "istio" {
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "istio"
      namespace = "argocd"
    }
    spec = {
      project = kubernetes_manifest.system_project.manifest.metadata.name
      sources = [{
        repoURL        = "https://istio-release.storage.googleapis.com/charts"
        targetRevision = local.istio_version
        chart          = "base"
        helm           = { releaseName = "istio" }
        }, {
        repoURL        = "https://istio-release.storage.googleapis.com/charts"
        targetRevision = local.istio_version
        chart          = "istiod"
        helm = {
          releaseName = "istio"
          values = yamlencode({
            profile = "ambient"
            # TODO: Revisit resource sizing after measuring usage.
            resources = { requests = { cpu = "100m", memory = "128Mi" } }
          })
        }
        }, {
        repoURL        = "https://istio-release.storage.googleapis.com/charts"
        targetRevision = local.istio_version
        chart          = "cni"
        helm = {
          releaseName = "istio-cni"
          values = yamlencode({
            profile = "ambient"
            global  = { platform = "k3s" }
          })
        }
        }, {
        repoURL        = "https://istio-release.storage.googleapis.com/charts"
        targetRevision = local.istio_version
        chart          = "ztunnel"
        helm = {
          releaseName = "ztunnel"
          values = yamlencode({
            resources = { requests = { memory = "128Mi" } }
          })
        }
      }]
      destination = { name = "in-cluster", namespace = "istio-system" }
      ignoreDifferences = concat(local.ignore_differences, [
        for name in ["istio-validator-istio-system", "istiod-default-validator"] : {
          group             = "admissionregistration.k8s.io"
          kind              = "ValidatingWebhookConfiguration"
          name              = name
          jqPathExpressions = [".webhooks[].failurePolicy", ".webhooks[].clientConfig.caBundle"]
        }
      ])
      syncPolicy = merge(local.sync_policy, {
        automated = {
          enabled  = true
          prune    = true
          selfHeal = true
        }
        syncOptions = concat(local.sync_policy.syncOptions, ["RespectIgnoreDifferences=true"])
      })
    }
  }
  lifecycle {
    prevent_destroy = true
  }

  wait {
    fields = {
      "status.sync.status"   = "Synced"
      "status.health.status" = "Healthy"
    }
  }

  timeouts {
    create = "15m"
    update = "15m"
  }
}

resource "kubernetes_manifest" "istio_resources" {
  depends_on = [kubernetes_manifest.istio]
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "istio-resources"
      namespace = "argocd"
    }
    spec = {
      project = kubernetes_manifest.system_project.manifest.metadata.name
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
              strict-mtls = {
                forceRename = "strict-mtls"
                manifest = {
                  apiVersion = "security.istio.io/v1"
                  kind       = "PeerAuthentication"

                  # Mesh-wide default.
                  spec = {
                    mtls = {
                      mode = "STRICT"
                    }
                  }
                }
              }

              gateway = {
                forceRename = "gateway"
                manifest = {
                  apiVersion = "gateway.networking.k8s.io/v1"
                  kind       = "Gateway"
                  metadata = {
                    namespace   = "istio-system"
                    annotations = { "cert-manager.io/cluster-issuer" = "acme" }
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
                      hostname      = "*.${var.domain}"
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
                    annotations = { "external-dns.kubernetes.io/target" = "${var.resource_prefix}-tunnel.${var.domain}" }
                  }
                  spec = {
                    gatewayClassName = "istio"
                    infrastructure   = { parametersRef = { group = "", kind = "ConfigMap", name = "public-gateway" } }
                    listeners = [{
                      name          = "https"
                      hostname      = "*.${var.domain}"
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
      destination       = { name = "in-cluster", namespace = "istio-system" }
      ignoreDifferences = local.ignore_differences
      syncPolicy = merge(local.sync_policy, {
        automated = {
          enabled  = true
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
