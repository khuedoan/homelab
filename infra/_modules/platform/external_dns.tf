resource "kubectl_manifest" "external_dns" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "external-dns"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://kubernetes-sigs.github.io/external-dns"
        chart          = "external-dns"
        targetRevision = "1.23.0"
        helm = {
          releaseName = "external-dns"
          values = yamlencode({
            # TODO: Revisit resource sizing after measuring usage.
            resources = { requests = { cpu = "10m", memory = "32Mi" } }
            env = [{
              name = "CF_API_TOKEN"
              valueFrom = {
                secretKeyRef = {
                  key  = "value"
                  name = "cloudflare-api-token"
                }
              }
            }]
            extraArgs          = ["--annotation-filter=external-dns.alpha.kubernetes.io/exclude notin (true)"]
            interval           = "5m"
            policy             = "upsert-only"
            serviceMonitor     = { enabled = true }
            provider           = { name = "cloudflare" }
            sources            = ["service", "gateway-httproute"]
            triggerLoopOnEvent = true
            txtOwnerId         = "homelab"
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "external-dns" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
