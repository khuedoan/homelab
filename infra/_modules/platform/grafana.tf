resource "kubectl_manifest" "grafana" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "grafana"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = var.repository.url
        targetRevision = var.repository.revision
        path           = "infra/_modules/platform/grafana"
        helm = {
          releaseName = "grafana"
          values = yamlencode({
            grafana = {
              sidecar = {
                # TODO: Revisit resource sizing after measuring usage.
                resources = { requests = { cpu = "10m", memory = "32Mi" } }
                dashboards = {
                  enabled         = true
                  searchNamespace = "monitoring-system"
                }
                datasources = {
                  enabled         = true
                  searchNamespace = "monitoring-system"
                }
              }
              # TODO: Revisit resource sizing after measuring usage.
              resources = {
                limits = {
                  memory = "512Mi"
                }
                requests = {
                  cpu    = "50m"
                  memory = "128Mi"
                }
              }
              "grafana.ini" = {
                "auth.generic_oauth" = {
                  allow_sign_up = true
                  api_url       = "https://auth.khuedoan.com/oauth2/openid/grafana/userinfo"
                  auth_url      = "https://auth.khuedoan.com/ui/oauth2"
                  client_id     = "grafana"
                  client_secret = "$__env{GRAFANA_SSO_CLIENT_SECRET}"
                  enabled       = true
                  name          = "Kanidm"
                  scopes        = "openid profile email groups"
                  token_url     = "https://auth.khuedoan.com/oauth2/token"
                  use_pkce      = true
                }
                server = {
                  root_url = "https://grafana.khuedoan.com"
                }
              }
              envFromSecret = "grafana-secrets"
              route = {
                main = {
                  enabled    = true
                  apiVersion = "gateway.networking.k8s.io/v1"
                  kind       = "HTTPRoute"
                  parentRefs = [{
                    name        = "gateway"
                    namespace   = "istio-system"
                    sectionName = "https"
                  }]
                  hostnames = ["grafana.khuedoan.com"]
                  matches   = [{ path = { type = "PathPrefix", value = "/" } }]
                }
              }
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "grafana" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
