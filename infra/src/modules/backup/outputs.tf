output "location" {
  value = {
    bucket   = cloudflare_r2_bucket.backup.name
    endpoint = "https://${var.cloudflare_account_id}.r2.cloudflarestorage.com"
  }
}

output "secrets" {
  sensitive = true
  value = {
    "velero/encryption" = {
      password = random_password.backup_repository.result
    }
    "velero/repository" = {
      credentials = <<-EOF
        [default]
        aws_access_key_id = ${cloudflare_account_token.backup.id}
        aws_secret_access_key = ${sha256(cloudflare_account_token.backup.value)}
      EOF
    }
  }
}
