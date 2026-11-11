resource "kubernetes_namespace_v1" "velero" {
  metadata {
    name = "velero"
  }
}

# Keep backup credentials independent of OpenBao so Velero can restore it
resource "kubernetes_secret_v1" "velero_credentials" {
  for_each = toset(["velero-credentials", "velero-repo-credentials"])

  metadata {
    name      = each.key
    namespace = kubernetes_namespace_v1.velero.metadata[0].name
  }

  data = {
    velero-credentials = {
      cloud = var.backup.credentials.repository_credentials
    }
    velero-repo-credentials = {
      repository-password = var.backup.credentials.repository_password
    }
  }[each.key]
}

resource "kubernetes_manifest" "velero" {
  depends_on = [kubernetes_secret_v1.velero_credentials]
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name        = "velero"
      namespace   = "argocd"
      annotations = { "argocd.argoproj.io/compare-options" = "IncludeMutationWebhook=true" }
    }
    spec = {
      project = kubernetes_manifest.system_project.manifest.metadata.name
      sources = [{
        repoURL        = "ghcr.io/vmware-tanzu/charts"
        chart          = "velero"
        targetRevision = "12.2.4"
        helm = {
          values = yamlencode({
            initContainers = [{
              name         = "aws-plugin"
              image        = "velero/velero-plugin-for-aws:v1.14.4"
              volumeMounts = [{ name = "plugins", mountPath = "/target" }]
            }]
            deployNodeAgent  = true
            nodeAgent        = { disableHostPath = true }
            snapshotsEnabled = false # Ceph CSI snapshots, not AWS EBS snapshots.
            configuration = {
              features = "EnableCSI"
              backupStorageLocation = [{
                provider   = "aws"
                bucket     = var.backup.location.bucket
                accessMode = var.recovery_mode ? "ReadOnly" : "ReadWrite"
                config = {
                  region            = var.backup.location.region
                  s3Url             = var.backup.location.endpoint
                  checksumAlgorithm = ""
                }
              }]
            }
            credentials = { existingSecret = "velero-credentials" }
            schedules = {
              cluster = {
                disabled = var.recovery_mode
                schedule = "0 * * * *"
                template = {
                  snapshotMoveData = true
                  excludedNamespaces = [
                    "argocd",
                    "grafana",
                    "istio-system",
                    "loki",
                    "monitoring-system",
                    "velero",
                  ]
                }
              }
            }
            metrics = { serviceMonitor = { enabled = true } }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "velero" }
      ignoreDifferences = local.ignore_differences
      syncPolicy = merge(local.sync_policy, {
        automated = {
          enabled  = true
          prune    = true
          selfHeal = true
        }
      })
    }
  }
  lifecycle {
    prevent_destroy = true
  }
}
