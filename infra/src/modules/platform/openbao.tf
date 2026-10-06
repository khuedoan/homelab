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
        repoURL        = "ghcr.io/bank-vaults/helm-charts"
        targetRevision = "1.24.1"
        chart          = "vault-operator"
        helm = {
          releaseName = "openbao"
          values      = yamlencode({ watchNamespace = "openbao" })
        }
        }, {
        repoURL        = "ghcr.io/bank-vaults/helm-charts"
        targetRevision = "0.4.1"
        chart          = "secrets-webhook"
        helm = {
          releaseName = "openbao"
          values = yamlencode({
            autoscaling = { hpa = { enabled = false } }
            resources   = { requests = { cpu = "10m", memory = "32Mi" } }
            env = {
              BAO_ADDR                     = "http://openbao.openbao.svc.cluster.local:8200"
              BAO_PATH                     = "kubernetes"
              BAO_ADDR_ALLOWLIST           = "http://openbao.openbao.svc.cluster.local:8200"
              BAO_ALLOW_OBJECT_SKIP_VERIFY = "false"
            }
            secretsFailurePolicy = "Fail"
            secrets              = { objectSelector = { matchLabels = { "homelab.khuedoan.com/bao-secret" = "true" } } }
            pods                 = { objectSelector = { matchLabels = { "homelab.khuedoan.com/bao-pod" = "true" } } }
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

