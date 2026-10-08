include "root" {
  path = find_in_parent_folders("root.hcl")
}

exclude {
  if      = true
  actions = ["all"]
  no_run  = true
}

terraform {
  source = "${find_in_parent_folders("_modules")}//cloudflare"
}

inputs = {
  resource_prefix = "homelab-staging"
  domain          = "staging.khuedoan.com"
  zone_name       = "khuedoan.com"
}
