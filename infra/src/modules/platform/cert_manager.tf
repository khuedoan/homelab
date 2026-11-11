resource "kubernetes_manifest" "cert_manager" {
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "cert-manager"
      namespace = "argocd"
    }
    spec = {
      project = kubernetes_manifest.system_project.manifest.metadata.name
      sources = [{
        repoURL        = "https://charts.jetstack.io"
        chart          = "cert-manager"
        targetRevision = "v1.21.2"
        helm = {
          releaseName = "cert-manager"
          values = yamlencode({
            crds = { enabled = true }
            config = {
              apiVersion = "controller.config.cert-manager.io/v1alpha1"
              kind       = "ControllerConfiguration"
              gatewayAPI = { enabled = true }
            }
            # TODO: Revisit resource sizing after measuring usage.
            resources       = { requests = { cpu = "10m", memory = "64Mi" } }
            webhook         = { resources = { requests = { cpu = "10m", memory = "32Mi" } } }
            cainjector      = { resources = { requests = { cpu = "10m", memory = "64Mi" } } }
            startupapicheck = { resources = { requests = { cpu = "10m", memory = "32Mi" } } }
            prometheus = {
              enabled = true
              servicemonitor = {
                enabled = true
              }
            }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "cert-manager" }
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

resource "kubernetes_manifest" "cert_manager_self_signed" {
  depends_on = [kubernetes_manifest.cert_manager]
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "cert-manager-self-signed"
      namespace = "argocd"
    }
    spec = {
      project = kubernetes_manifest.system_project.manifest.metadata.name
      sources = [{
        repoURL        = "https://bjw-s-labs.github.io/helm-charts"
        chart          = "app-template"
        targetRevision = "5.2.1"
        helm = {
          releaseName = "cert-manager-self-signed"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            rawResources = {
              self-signed = {
                forceRename = "self-signed"
                manifest = {
                  apiVersion = "cert-manager.io/v1"
                  kind       = "ClusterIssuer"
                  spec       = { selfSigned = {} }
                }
              }
            }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "cert-manager" }
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

resource "kubernetes_manifest" "cert_manager_acme" {
  depends_on = [kubernetes_manifest.openbao]
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name        = "cert-manager-acme"
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
          releaseName = "cert-manager-acme"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            secrets = {
              cloudflare = {
                forceRename = "cloudflare-api-token"
                stringData  = { "api-token" = "bao:secret/data/cert-manager/cloudflare#api_token" }
              }
            }
            rawResources = {
              acme = {
                forceRename = "acme"
                manifest = {
                  apiVersion = "cert-manager.io/v1"
                  kind       = "ClusterIssuer"
                  spec = {
                    acme = {
                      server              = "https://acme-v02.api.letsencrypt.org/directory"
                      privateKeySecretRef = { name = "acme" }
                      solvers = [{
                        selector = { dnsZones = [var.domain] }
                        dns01 = {
                          cloudflare = {
                            apiTokenSecretRef = { name = "cloudflare-api-token", key = "api-token" }
                          }
                        }
                      }]
                    }
                  }
                }
              }
            }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "cert-manager" }
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
