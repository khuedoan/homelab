output "location" {
  value = {
    bucket   = b2_bucket.backup.bucket_name
    endpoint = data.b2_account_info.backup.s3_api_url
    region   = split(".", data.b2_account_info.backup.s3_api_url)[1]
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
        aws_access_key_id = ${b2_application_key.backup.application_key_id}
        aws_secret_access_key = ${b2_application_key.backup.application_key}
      EOF
    }
  }
}
