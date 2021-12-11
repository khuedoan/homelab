provider "b2" {
  # Environment variables
  # B2_APPLICATION_KEY
  # B2_APPLICATION_KEY_ID
}

resource "b2_bucket" "backup" {
  bucket_name = "khuedoan-homelab-backup"
  bucket_type = "allPrivate"
}
