resource "kubectl_manifest" "kiali" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "kiali"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://kiali.org/helm-charts"
        targetRevision = "2.33.0"
        chart          = "kiali-server"
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
      destination       = { server = "https://kubernetes.default.svc", namespace = "kiali" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}

resource "kubectl_manifest" "kiali_resources" {
  depends_on        = [kubectl_manifest.openbao]
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "kiali-resources"
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
          releaseName = "kiali"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            secrets = {
              main = {
                labels = { "homelab.khuedoan.com/bao-secret" = "true" }
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
      destination       = { server = "https://kubernetes.default.svc", namespace = "kiali" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
