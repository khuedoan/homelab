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
        repoURL        = var.repository.url
        targetRevision = var.repository.revision
        path           = "infra/_modules/platform/cert-manager"
        helm = {
          releaseName = "cert-manager"
          values = yamlencode({
            acme = {
              enabled = true
            }
            cert-manager = {
              installCRDs = true
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
