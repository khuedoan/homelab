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
    key                         = "${local.state_key}/tfstate.json"
    region                      = "auto"
    endpoints                   = { s3 = "https://${local.r2_account_id}.r2.cloudflarestorage.com" }
    use_lockfile                = true
    profile                     = "r2-state"
    shared_config_files         = ["./r2-state.config"]
    skip_credentials_validation = true
    skip_region_validation      = true
    skip_requesting_account_id  = true
    skip_metadata_api_check     = true
    skip_s3_checksum            = true
  }
}

generate "r2_credentials" {
  path      = "r2-state.config"
  if_exists = "overwrite_terragrunt"
  contents  = <<-EOF
    [profile r2-state]
    credential_process = lab infra state credentials --account-id ${local.r2_account_id} --bucket ${local.state_bucket}
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
      "lab", "cluster", "kubeconfig",
      "--environment", local.env,
      "--output", local.kubeconfig,
    ]
  }

  before_hook "bootstrap_tfstate" {
    commands = ["init"]
    execute = [
      "lab", "infra", "state", "ensure",
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
