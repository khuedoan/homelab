locals {
  env        = split("/", path_relative_to_include())[0]
  unit       = basename(get_original_terragrunt_dir())
  state_key  = trimprefix(path_relative_to_include(), "${local.env}/")
  kubeconfig = "${get_repo_root()}/infra/${local.env}/kubeconfig.yaml"
  identity = jsondecode(run_cmd(
    "--terragrunt-global-cache", "--terragrunt-quiet",
    "env", "-u", "CLOUDFLARE_API_TOKEN", "-u", "CLOUDFLARE_API_KEY", "-u", "CLOUDFLARE_EMAIL",
    "CF_SEND_TELEMETRY=false", "cf", "--quiet", "auth", "whoami",
  ))
  access_token  = jsondecode(file(trimprefix(local.identity.authSource, "OAuth token from "))).oauth_token
  r2_account_id = get_env("CLOUDFLARE_ACCOUNT_ID", local.identity.accounts[0].id)
  state_bucket  = "homelab-${local.env}-tfstate"
}

# Native S3 lockfiles require OpenTofu >= 1.10. Use generate instead of
# remote_state so Terragrunt does not try to manage R2 as an AWS S3 bucket.
# Credentials come from the AWS SDK environment/profile, never backend.tf.
terraform_binary             = "tofu"
terraform_version_constraint = ">= 1.10.0"

generate "backend" {
  path      = "backend.tf"
  if_exists = "overwrite_terragrunt"
  contents  = <<-EOF
    terraform {
      required_version = ">= 1.10.0"

      backend "s3" {
        bucket       = "${local.state_bucket}"
        key          = "${local.state_key}/tfstate.json"
        region       = "auto"
        use_lockfile = true

        endpoints = {
          s3 = "https://${local.r2_account_id}.r2.cloudflarestorage.com"
        }

        use_path_style              = true
        skip_credentials_validation = true
        skip_metadata_api_check     = true
        skip_region_validation      = true
        skip_requesting_account_id  = true
        skip_s3_checksum            = true
      }
    }
  EOF
}

terraform {
  extra_arguments "cloudflare_oauth" {
    commands = ["plan", "apply", "import", "refresh"]
    env_vars = {
      CLOUDFLARE_API_TOKEN = local.access_token
      CLOUDFLARE_API_KEY   = ""
      CLOUDFLARE_EMAIL     = ""
    }
  }

  before_hook "verify_cluster" {
    commands    = ["plan", "apply", "refresh", "import"]
    if          = contains(["argocd", "platform"], local.unit)
    working_dir = get_repo_root()
    execute = [
      "toolbox", "cluster", "kubeconfig",
      "--environment", local.env,
      "--output", local.kubeconfig,
    ]
  }

  before_hook "bootstrap_tfstate" {
    commands = ["init"]
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

inputs = {
  cloudflare_account_id = local.r2_account_id
  kubeconfig            = local.kubeconfig
}
