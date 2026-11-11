resource "kubernetes_manifest" "cloudflared" {
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "cloudflared"
      namespace = "argocd"
    }
    spec = {
      project = kubernetes_manifest.system_project.manifest.metadata.name
      sources = [{
        repoURL        = "https://bjw-s-labs.github.io/helm-charts"
        chart          = "app-template"
        targetRevision = "3.1.0"
        helm = {
          releaseName = "cloudflared"
          values = yamlencode({
            configMaps = {
              config = {
                data = {
                  "config.yaml" = yamlencode({
                    tunnel           = "${var.resource_prefix}-tunnel"
                    credentials-file = "/etc/cloudflared/credentials.json"
                    metrics          = "0.0.0.0:2000"
                    no-autoupdate    = true
                    ingress = [
                      {
                        hostname = "www.${var.domain}"
                        service  = "https://public-istio.istio-system"
                      },
                      {
                        hostname = "draw.${var.domain}"
                        service  = "https://public-istio.istio-system"
                      },
                      {
                        hostname = "chat.${var.domain}"
                        service  = "https://public-istio.istio-system"
                      },
                      {
                        hostname = "*.${var.domain}"
                        service  = "https://gateway-istio.istio-system"
                      },
                      { service = "http_status:404" },
                    ]
                    originRequest = { originServerName = "tunnel.${var.domain}" }
                  })
                }
                enabled = true
              }
            }
            controllers = {
              cloudflared = {
                containers = {
                  app = {
                    args = ["tunnel", "--config", "/etc/cloudflared/config.yaml", "run"]
                    image = {
                      repository = "docker.io/cloudflare/cloudflared"
                      tag        = "2026.10.3"
                    }
                    # TODO: Revisit resource sizing after measuring usage.
                    resources = { requests = { cpu = "10m", memory = "32Mi" } }
                  }
                }
              }
            }
            persistence = {
              config = {
                enabled = true
                globalMounts = [{
                  path    = "/etc/cloudflared/config.yaml"
                  subPath = "config.yaml"
                }]
                name = "cloudflared-config"
                type = "configMap"
              }
              credentials = {
                enabled = true
                globalMounts = [{
                  path    = "/etc/cloudflared/credentials.json"
                  subPath = "credentials.json"
                }]
                name = "cloudflared-credentials"
                type = "secret"
              }
            }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "cloudflared" }
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

resource "kubernetes_manifest" "cloudflared_resources" {
  depends_on = [kubernetes_manifest.openbao]
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name        = "cloudflared-resources"
      namespace   = "argocd"
      annotations = { "argocd.argoproj.io/compare-options" = "IncludeMutationWebhook=true" }
    }
    spec = {
      project = kubernetes_manifest.system_project.manifest.metadata.name
      sources = [{
        repoURL        = "ghcr.io/bjw-s-labs/helm"
        chart          = "app-template"
        targetRevision = "5.3.0"
        helm = {
          releaseName = "cloudflared-resources"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            secrets = {
              cloudflare = {
                forceRename = "cloudflared-credentials"
                stringData  = { "credentials.json" = "bao:secret/data/cloudflared/cloudflare#credentials.json" }
              }
            }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "cloudflared" }
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
