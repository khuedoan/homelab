resource "kubectl_manifest" "openbao" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "openbao"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = var.repository.url
        targetRevision = var.repository.revision
        path           = "infra/_modules/platform/openbao"
        helm = {
          releaseName = "openbao"
          values = yamlencode({
            readers = merge(yamldecode(file("${path.module}/openbao/values.yaml")).readers, {
              for name, app in var.apps : app.namespace => [name == "paperless" ? "paperless.admin" : "${name}.auth"]
              if contains(["paperless", "tailscale", "wireguard"], name)
            })
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "openbao" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
