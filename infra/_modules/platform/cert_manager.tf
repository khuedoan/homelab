resource "kubectl_manifest" "cert_manager" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "cert-manager"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://charts.jetstack.io"
        targetRevision = "v1.21.2"
        chart          = "cert-manager"
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
      destination       = { server = "https://kubernetes.default.svc", namespace = "cert-manager" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
  wait_for {
    field {
      key   = "status.sync.status"
      value = "Synced"
    }
    field {
      key   = "status.health.status"
      value = "Healthy"
    }
  }
  timeouts {
    create = "15m"
    update = "15m"
  }
}

resource "kubectl_manifest" "cert_manager_resources" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "cert-manager-resources"
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
          releaseName = "cert-manager-resources"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            rawResources = {
              letsencrypt-prod = {
                forceRename = "letsencrypt-prod"
                manifest = {
                  apiVersion = "cert-manager.io/v1"
                  kind       = "ClusterIssuer"
                  spec = {
                    acme = {
                      server              = "https://acme-v02.api.letsencrypt.org/directory"
                      privateKeySecretRef = { name = "letsencrypt-prod" }
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
      destination       = { server = "https://kubernetes.default.svc", namespace = "cert-manager" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
