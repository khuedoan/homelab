resource "b2_bucket" "backup" {
  bucket_name = "khuedoan-homelab-backup"
  bucket_type = "allPrivate"
}

resource "random_password" "backup_repo_password" {
  length  = 64
  special = false
}

resource "kubernetes_secret" "backblaze_credentials" {
  metadata {
    name      = "backblaze-credentials"
    namespace = "k8up-operator"
  }

  data = {
    "application-key-id" = var.b2_application_key_id
    "application-key"    = var.b2_application_key
    "repo-password"      = random_password.backup_repo_password.result
  }
}
