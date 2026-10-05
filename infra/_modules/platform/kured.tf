resource "kubectl_manifest" "kured" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "kured"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://kubereboot.github.io/charts"
        chart          = "kured"
        targetRevision = "4.7.0"
        helm = {
          releaseName = "kured"
          values = yamlencode({
            # TODO: Revisit resource sizing after measuring usage.
            resources = { requests = { cpu = "10m", memory = "32Mi" } }
            configuration = {
              annotateNodes         = true
              rebootSentinelCommand = "sh -c \"! needs-restarting --reboothint\""
              timeZone              = "Asia/Ho_Chi_Minh"
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "kured" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
