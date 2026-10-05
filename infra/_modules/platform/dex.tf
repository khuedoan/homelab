resource "kubectl_manifest" "dex" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "dex"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = var.repository.url
        targetRevision = var.repository.revision
        path           = "infra/_modules/platform/dex"
        helm = {
          releaseName = "dex"
          values = yamlencode({
            dex = {
              # TODO: Revisit resource sizing after measuring usage.
              resources = { requests = { cpu = "10m", memory = "32Mi" } }
              config = {
                connectors = [{
                  config = {
                    clientID     = "$KANIDM_CLIENT_ID"
                    clientSecret = "$KANIDM_CLIENT_SECRET"
                    issuer       = "https://auth.khuedoan.com/oauth2/openid/dex"
                    redirectURI  = "https://dex.khuedoan.com/callback"
                    scopes       = ["openid", "profile", "email", "groups"]
                  }
                  id   = "kanidm"
                  name = "Kanidm"
                  type = "oidc"
                }]
                issuer = "https://dex.khuedoan.com"
                oauth2 = {
                  skipApprovalScreen = true
                }
                staticClients = [{
                  id           = "grafana-sso"
                  name         = "Grafana"
                  redirectURIs = ["https://grafana.khuedoan.com/login/generic_oauth"]
                  secretEnv    = "GRAFANA_SSO_CLIENT_SECRET"
                  }, {
                  id           = "forgejo"
                  name         = "Forgejo"
                  redirectURIs = ["https://git.khuedoan.com/user/oauth2/Dex/callback"]
                  secretEnv    = "FORGEJO_CLIENT_SECRET"
                }]
                storage = {
                  config = {
                    inCluster = true
                  }
                  type = "kubernetes"
                }
              }
              envFrom = [{
                secretRef = {
                  name = "dex-secrets"
                }
              }]
              ingress = {
                annotations = {
                  "cert-manager.io/cluster-issuer" = "letsencrypt-prod"
                }
                className = "nginx"
                enabled   = true
                hosts = [{
                  host = "dex.khuedoan.com"
                  paths = [{
                    path     = "/"
                    pathType = "ImplementationSpecific"
                  }]
                }]
                tls = [{
                  hosts      = ["dex.khuedoan.com"]
                  secretName = "dex-tls-certificate"
                }]
              }
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "dex" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
