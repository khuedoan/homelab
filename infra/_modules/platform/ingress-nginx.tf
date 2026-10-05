resource "kubectl_manifest" "ingress_nginx" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "ingress-nginx"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://kubernetes.github.io/ingress-nginx"
        chart          = "ingress-nginx"
        targetRevision = "4.11.2"
        helm = {
          releaseName = "ingress-nginx"
          values = yamlencode({
            controller = {
              # TODO: Revisit resource sizing after measuring usage.
              resources = { requests = { cpu = "50m", memory = "64Mi" } }
              metrics = {
                enabled = true
                serviceMonitor = {
                  enabled = true
                }
              }
              admissionWebhooks = {
                timeoutSeconds = 30
              }
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "ingress-nginx" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
