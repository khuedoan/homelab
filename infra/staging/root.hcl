locals {
  env           = basename(get_parent_terragrunt_dir())
  r2_access_key = get_env("CLOUDFLARE_TFSTATE_ACCESS_KEY", "")
  r2_secret_key = get_env("CLOUDFLARE_TFSTATE_SECRET_KEY", "")
  r2_account_id = get_env("CLOUDFLARE_ACCOUNT_ID", "")
  # Metal only writes a local file and converges on every run, so it keeps local
  # state instead of the shared S3 backend.
  stateless = path_relative_to_include() == "metal"
}

generate "backend" {
  path              = "backend.tf.json"
  if_exists         = "overwrite"
  disable           = local.stateless
  if_disabled       = "remove_terragrunt"
  disable_signature = true
  contents = jsonencode({
    terraform = {
      backend = {
        s3 = {
          bucket                      = "tfstate-${local.env}"
          key                         = "${path_relative_to_include()}/tfstate.json"
          region                      = "auto"
          skip_credentials_validation = true
          skip_metadata_api_check     = true
          skip_region_validation      = true
          skip_requesting_account_id  = true
          skip_s3_checksum            = true
          use_path_style              = true
          access_key                  = local.r2_access_key
          secret_key                  = local.r2_secret_key
          endpoints = {
            s3 = "https://${local.r2_account_id}.r2.cloudflarestorage.com"
          }
        }
      }
    }
  })
}

terraform {
  before_hook "unsupported_destroy" {
    commands = ["destroy"]
    execute  = ["sh", "-c", "echo 'Infrastructure teardown is unsupported' >&2; exit 1"]
  }
}
