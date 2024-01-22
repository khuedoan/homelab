resource "kubernetes_secret" "velero_credentials" {
  metadata {
    name      = "velero-credentials"
    namespace = "velero"
  }

  data = {
    cloud = <<EOF
      %{for bucket in var.backup_buckets~}
      [${bucket.name}]
      aws_access_key_id=${bucket.access_key_id}
      aws_secret_access_key=${bucket.secret_access_key}
      %{endfor~}
    EOF
  }
}
