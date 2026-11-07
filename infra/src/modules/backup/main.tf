terraform {
  required_providers {
    b2 = {
      source  = "Backblaze/b2"
      version = "= 0.14.1"
    }

    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}

provider "b2" {}

data "b2_account_info" "backup" {}

resource "random_password" "backup_repository" {
  length = 32

  lifecycle {
    prevent_destroy = true
  }
}

resource "b2_bucket" "backup" {
  bucket_name = "${var.resource_prefix}-backup-${data.b2_account_info.backup.account_id}"
  bucket_type = "allPrivate"

  lifecycle_rules {
    file_name_prefix                                       = ""
    days_from_hiding_to_deleting                           = 1
    days_from_starting_to_canceling_unfinished_large_files = 1
  }

  lifecycle {
    prevent_destroy = true
  }
}

resource "b2_application_key" "backup" {
  key_name   = "${var.resource_prefix}-backup"
  bucket_ids = [b2_bucket.backup.bucket_id]
  capabilities = [
    "listBuckets", "listAllBucketNames", "listFiles", "readFiles", "writeFiles", "deleteFiles",
  ]
}
