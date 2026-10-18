resource "kubectl_manifest" "velero" {
  depends_on        = [kubectl_manifest.openbao]
  server_side_apply = true
  yaml_body = yamlencode({
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "velero"
      namespace = "argocd"
      labels    = { "app.kubernetes.io/part-of" = "homelab" }
    }
    spec = {
      project = "default"
      sources = [{
        repoURL        = "https://vmware-tanzu.github.io/helm-charts"
        chart          = "velero"
        targetRevision = "12.2.1"
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
                bucket     = var.backup_location.bucket
                accessMode = var.backup_read_only ? "ReadOnly" : "ReadWrite"
                config = {
                  region            = "auto"
                  s3Url             = var.backup_location.endpoint
                  checksumAlgorithm = ""
                }
              }]
            }
            credentials = { existingSecret = "velero-credentials" }
            extraObjects = [for name, data in {
              velero-credentials = { cloud = "bao:secret/data/velero/repository#credentials" }
              velero-repo-credentials = {
                repository-password = "bao:secret/data/velero/encryption#password"
              }
              } : {
              apiVersion = "v1"
              kind       = "Secret"
              metadata = {
                name        = name
                labels      = { "homelab.khuedoan.com/bao-secret" = "true" }
                annotations = { "argocd.argoproj.io/sync-wave" = "-1" }
              }
              stringData = data
            }]
            schedules = {
              cluster = {
                disabled = var.backup_read_only
                schedule = "0 * * * *"
                template = { snapshotMoveData = true }
              }
            }
            metrics = { serviceMonitor = { enabled = true } }
          })
        }
      }]
      destination       = { name = "in-cluster", namespace = "velero" }
      ignoreDifferences = local.ignore_differences
      syncPolicy        = local.sync_policy
    }
  })
  lifecycle {
    prevent_destroy = true
  }
}
