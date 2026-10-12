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
        repoURL        = "https://grafana-community.github.io/helm-charts"
        targetRevision = "13.5.0"
        chart          = "grafana"
        helm = {
          releaseName = "grafana"
          values = yamlencode({
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
            resources = {
              limits = {
                memory = "1Gi"
              }
              requests = {
                cpu    = "50m"
                memory = "256Mi"
              }
            }
            "grafana.ini" = {
              auth = { disable_login_form = true }
              "auth.generic_oauth" = {
                allow_assign_grafana_admin = true
                # Kanidm's groups scope emits UUIDs and SPNs, not bare names.
                role_attribute_path        = "contains(groups[*], 'grafana_admins@auth.${var.domain}') && 'GrafanaAdmin' || 'Viewer'"
                api_url                    = "https://auth.${var.domain}/oauth2/openid/grafana/userinfo"
                auth_url                   = "https://auth.${var.domain}/ui/oauth2"
                client_id                  = "grafana"
                client_secret              = "$__env{GRAFANA_SSO_CLIENT_SECRET}"
                enabled                    = true
                name                       = "SSO"
                scopes                     = "openid profile email groups"
                token_url                  = "https://auth.${var.domain}/oauth2/token"
                use_pkce                   = true
              }
              server = {
                root_url = "https://grafana.${var.domain}"
              }
            }
            envFromSecret = "grafana-secrets"
            route = {
              main = {
                enabled = true
                parentRefs = [{
                  name        = "gateway"
                  namespace   = "istio-system"
                  sectionName = "https"
                }]
                hostnames = ["grafana.${var.domain}"]
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

resource "kubectl_manifest" "grafana_resources" {
  depends_on        = [kubectl_manifest.openbao]
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "grafana-resources"
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
          releaseName = "grafana-resources"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            secrets = {
              grafana-secrets = {
                forceRename = "grafana-secrets"
                labels      = { "homelab.khuedoan.com/bao-secret" = "true" }
                stringData = {
                  GRAFANA_SSO_CLIENT_SECRET = "bao:secret/data/grafana/sso#client_secret"
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
