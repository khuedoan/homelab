variable "backup_buckets" {
  type = list(object({
    name              = string
    url               = string
    bucket            = string
    region            = string
    access_key_id     = string
    secret_access_key = string
  }))
}
