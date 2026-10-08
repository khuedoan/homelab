include "root" {
  path = find_in_parent_folders("root.hcl")
}

include "kubernetes" {
  path = find_in_parent_folders("kubernetes.hcl")
}

dependencies {
  paths = ["../cluster"]
}

terraform {
  source = "${find_in_parent_folders("_modules")}//argocd"
}

inputs = {
  domain = "khuedoan.com"
}
