locals {
  root          = read_terragrunt_config(find_in_parent_folders("root.hcl"))
  auth_commands = ["plan", "apply", "import", "refresh"]
  access_token  = local.root.locals.access_token
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

inputs = {
  cloudflare_account_id = local.root.locals.r2_account_id
}
