include "root" {
  path   = find_in_parent_folders("root.hcl")
  expose = true
}

locals {
  api_token  = get_env("CLOUDFLARE_TFSTATE_API_TOKEN", "")
  account_id = get_env("CLOUDFLARE_ACCOUNT_ID", "")
  bucket     = get_env("TFSTATE_BUCKET", "tfstate-${include.root.locals.env}")
}

terraform {
  source = "${find_in_parent_folders("_modules")}//tfstate"

  before_hook "bootstrap_tfstate" {
    commands = ["init", "plan", "apply"]
    execute = [
      "go", "run", ".",
      "--api-token=${local.api_token}",
      "--account-id=${local.account_id}",
      "--bucket=${local.bucket}",
    ]
  }
}
