include "root" {
  path = find_in_parent_folders("root.hcl")
}

include "kubernetes" {
  path = find_in_parent_folders("_kubernetes.hcl")
}

exclude {
  if      = true
  actions = ["all"]
  no_run  = true
}

dependencies {
  paths = ["../cluster"]
}

terraform {
  source = "${find_in_parent_folders("_modules")}//cloudflare"
}
