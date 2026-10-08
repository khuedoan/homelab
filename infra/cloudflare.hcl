locals {
  auth_commands = ["plan", "apply", "import", "refresh"]
  needs_auth    = contains(local.auth_commands, get_terraform_command())
  identity = local.needs_auth ? jsondecode(run_cmd(
    "--terragrunt-quiet",
    "env", "-u", "CLOUDFLARE_API_TOKEN", "-u", "CLOUDFLARE_API_KEY", "-u", "CLOUDFLARE_EMAIL",
    "CF_SEND_TELEMETRY=false", "cf", "--quiet", "auth", "whoami",
  )) : null
  access_token = local.needs_auth ? jsondecode(file(trimprefix(local.identity.authSource, "OAuth token from "))).oauth_token : ""
}

terraform {
  source = "${find_in_parent_folders("_modules")}//cloudflare"

  extra_arguments "cloudflare_oauth" {
    commands = local.auth_commands
    env_vars = {
      CLOUDFLARE_API_TOKEN = local.access_token
      CLOUDFLARE_API_KEY   = ""
      CLOUDFLARE_EMAIL     = ""
    }
  }
}
