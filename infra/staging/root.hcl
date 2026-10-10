locals {
  env = basename(get_parent_terragrunt_dir())
  identity = jsondecode(run_cmd(
    "--terragrunt-global-cache", "--terragrunt-quiet",
    "env", "-u", "CLOUDFLARE_API_TOKEN", "-u", "CLOUDFLARE_API_KEY", "-u", "CLOUDFLARE_EMAIL",
    "CF_SEND_TELEMETRY=false", "cf", "--quiet", "auth", "whoami",
  ))
  access_token  = jsondecode(file(trimprefix(local.identity.authSource, "OAuth token from "))).oauth_token
  r2_account_id = get_env("CLOUDFLARE_ACCOUNT_ID", local.identity.accounts[0].id)
  state_bucket  = "homelab-${local.env}-tfstate" # TODO rename to homelab-tfstate-$env
  # Metal and cluster only write local files and converge on every run, so
  # they keep local state instead of the shared R2 backend.
  stateless = contains(["metal", "cluster"], path_relative_to_include())
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
        http = {
          address       = "https://api.cloudflare.com/client/v4/accounts/${local.r2_account_id}/r2/buckets/${local.state_bucket}/objects/${path_relative_to_include()}/tfstate.json"
          update_method = "PUT"
          headers       = { Authorization = "Bearer ${local.access_token}" }
        }
      }
    }
  })
}

terraform {
  extra_arguments "backend_credentials" {
    commands  = ["init"]
    arguments = ["-reconfigure"]
  }

  before_hook "bootstrap_tfstate" {
    commands = ["init", "plan", "apply"]
    execute = [
      "toolbox", "infra", "state", "ensure",
      "--account-id", local.r2_account_id,
      "--bucket", local.state_bucket,
    ]
  }

  before_hook "unsupported_destroy" {
    commands = ["destroy"]
    execute  = ["sh", "-c", "echo 'Infrastructure teardown is unsupported' >&2; exit 1"]
  }
}
