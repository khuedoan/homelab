resource "kubectl_manifest" "renovate" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "renovate"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://docs.renovatebot.com/helm-charts"
        targetRevision = "31.97.3"
        chart          = "renovate"
        helm = {
          releaseName = "renovate"
          values = yamlencode({
            # TODO: Revisit resource sizing after measuring usage.
            resources = { requests = { cpu = "50m", memory = "128Mi" } }
            cronjob = {
              schedule = "0 9 * * *"
            }
            existingSecret = "renovate-secret"
            renovate = {
              config = <<-EOT
{
  "platform": "gitea",
  "endpoint": "https://git.khuedoan.com/api/v1",
  "gitAuthor": "Renovate Bot <bot@renovateapp.com>",
  "autodiscover": true
}
EOT

            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "renovate" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}

resource "kubectl_manifest" "renovate_resources" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "renovate-resources"
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
          releaseName = "renovate-resources"
          values = yamlencode({
            global = { createDefaultServiceAccount = false }
            rawResources = {
              renovate-secret = {
                forceRename = "renovate-secret"
                manifest = {
                  apiVersion = "v1"
                  kind       = "Secret"
                  metadata = {
                    namespace = "renovate"
                    labels    = { "homelab.khuedoan.com/bao-secret" = "true" }
                    annotations = {
                      "secrets-webhook.security.bank-vaults.io/provider"           = "bao"
                      "secrets-webhook.security.bank-vaults.io/bao-role"           = "renovate"
                      "secrets-webhook.security.bank-vaults.io/bao-path"           = "kubernetes"
                      "secrets-webhook.security.bank-vaults.io/bao-serviceaccount" = "default"
                    }
                  }
                  data = {
                    RENOVATE_TOKEN = base64encode("bao:secret/data/forgejo.renovate#token")
                  }
                }
              }
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "renovate" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
