include "root" {
  path = find_in_parent_folders("root.hcl")
}

terraform {
  source = "${find_in_parent_folders("_modules")}//cloudflare"
}

inputs = {
  resource_prefix = "homelab-production"
  domain          = "khuedoan.com"
  zone_name       = "khuedoan.com"
}
