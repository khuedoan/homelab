include "root" {
  path = find_in_parent_folders("root.hcl")
}

terraform {
  source = "${find_in_parent_folders("_modules")}//backup"
}

inputs = {
  resource_prefix = "homelab-production"
}
