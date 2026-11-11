resource "kubernetes_manifest" "kiali" {
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "kiali"
      namespace = "argocd"
    }
    spec = {
      project = kubernetes_manifest.system_project.manifest.metadata.name
      sources = [{
        repoURL        = "https://kiali.org/helm-charts"
        chart          = "kiali-server"
        targetRevision = "2.34.1"
        helm = {
          releaseName = "kiali"
          values = yamlencode({
            auth = {
              strategy = "openid"
              openid = {
                client_id      = "kiali"
                issuer_uri     = "https://auth.${var.domain}/oauth2/openid/kiali"
                disable_rbac   = true
                username_claim = "preferred_username"
              }
            }
            deployment  = { view_only_mode = true }
            login_token = { signing_key = "secret:kiali:signing-key" }
            external_services = {
              prometheus = { url = "http://monitoring-system-kube-pro-prometheus.monitoring-system:9090" }
              grafana    = { enabled = false }
              tracing    = { enabled = false }
            }
            server = {
              web_root   = "/"
              web_fqdn   = "kiali.${var.domain}"
              web_schema = "https"
              web_port   = 443
            }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "kiali" }
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

resource "kubernetes_manifest" "kiali_resources" {
  depends_on = [kubernetes_manifest.openbao]
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name        = "kiali-resources"
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
          releaseName = "kiali"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            secrets = {
              main = {
                stringData = {
                  oidc-secret = "bao:secret/data/kiali/sso#client_secret"
                  signing-key = "bao:secret/data/kiali/session#signing_key"
                }
              }
            }
            route = {
              main = {
                parentRefs = [{
                  name        = "gateway"
                  namespace   = "istio-system"
                  sectionName = "https"
                }]
                hostnames = ["kiali.${var.domain}"]
                rules = [{
                  backendRefs = [{ name = "kiali", port = 20001 }]
                }]
              }
            }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "kiali" }
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
