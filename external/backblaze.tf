resource "b2_bucket" "backup" {
  bucket_name = "khuedoan-homelab-backup"
  bucket_type = "allPrivate"
}
