locals {
  env          = split("/", path_relative_to_include())[0]
  unit         = basename(path_relative_to_include())
  account      = get_env("CLOUDFLARE_ACCOUNT_ID")
  state_bucket = "homelab-${local.env}-tfstate"
}

remote_state {
  backend                         = "s3"
  disable_init                    = true
  disable_dependency_optimization = true
  generate = {
    path      = "backend.tf"
    if_exists = "overwrite_terragrunt"
  }
  config = {
    bucket                      = local.state_bucket
    key                         = "${local.unit}/tfstate.json"
    region                      = "auto"
    endpoints                   = { s3 = "https://${local.account}.r2.cloudflarestorage.com" }
    use_lockfile                = true
    skip_credentials_validation = true
    skip_region_validation      = true
    skip_requesting_account_id  = true
    skip_metadata_api_check     = true
    skip_s3_checksum            = true
  }
}

errors {
  retry "r2_token_propagation" {
    retryable_errors   = [".*S3: ListObjectsV2.*StatusCode: 401.*"]
    max_attempts       = 6
    sleep_interval_sec = 10
  }
}

terraform {
  source = "${find_in_parent_folders("src")}//modules/${local.unit}"
}

inputs = {
  cloudflare_account_id = local.account
  resource_prefix       = "homelab-${local.env}"
}
