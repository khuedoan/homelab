resource "kubectl_manifest" "zot" {
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "zot"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = var.repository.url
        targetRevision = var.repository.revision
        path           = "infra/_modules/platform/zot"
        helm = {
          releaseName = "zot"
          values = yamlencode({
            zot = {
              # TODO: Revisit resource sizing after measuring usage.
              resources   = { requests = { cpu = "50m", memory = "128Mi" } }
              persistence = true
              pvc = {
                create  = true
                storage = "2Gi"
                # TODO: Replace temporary local-path storage with replicated storage.
                storageClassName = "local-path"
              }
              ingress = {
                annotations = {
                  "cert-manager.io/cluster-issuer"              = "letsencrypt-prod"
                  "nginx.ingress.kubernetes.io/proxy-body-size" = "0"
                }
                className = "nginx"
                enabled   = true
                hosts = [{
                  host = "registry.khuedoan.com"
                  paths = [{
                    path = "/"
                  }]
                }]
                tls = [{
                  hosts      = ["registry.khuedoan.com"]
                  secretName = "zot-tls-certificate"
                }]
              }
            }
          })
        }
      }]
      destination       = { server = "https://kubernetes.default.svc", namespace = "zot" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
