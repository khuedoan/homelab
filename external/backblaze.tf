resource "b2_bucket" "backup" {
  bucket_name = "homelab-backup"
  bucket_type = "allPrivate"
}
