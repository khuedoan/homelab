resource "b2_bucket" "backup" {
  bucket_name = "khuedoan-homelab-backup"
  bucket_type = "allPrivate"
}

resource "kubernetes_secret" "backblaze_credentials" {
  metadata {
    name      = "backblaze-credentials"
    namespace = "k8up-operator"
  }

  data = {
    "application-key-id" = var.b2_application_key_id
    "application-key"    = var.b2_application_key
  }
}
