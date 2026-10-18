terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "= 5.27.0"
    }

    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}

# TODO migrate backups to Backblaze B2
provider "cloudflare" {}

data "cloudflare_account_api_token_permission_groups_list" "backup" {
  account_id = var.cloudflare_account_id
}

resource "random_password" "backup_repository" {
  length = 32

  lifecycle {
    prevent_destroy = true
  }
}

resource "cloudflare_r2_bucket" "backup" {
  account_id = var.cloudflare_account_id
  name       = "${var.resource_prefix}-backup"

  lifecycle {
    prevent_destroy = true
  }
}

resource "cloudflare_account_token" "backup" {
  account_id = var.cloudflare_account_id
  name       = cloudflare_r2_bucket.backup.name
  policies = [{
    effect = "allow"
    permission_groups = [
      for group in data.cloudflare_account_api_token_permission_groups_list.backup.result : { id = group.id }
      if group.name == "Workers R2 Storage Bucket Item Write"
    ]
    resources = jsonencode({
      "com.cloudflare.edge.r2.bucket.${var.cloudflare_account_id}_default_${cloudflare_r2_bucket.backup.name}" = "*"
    })
  }]
}
