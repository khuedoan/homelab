include "root" {
  path = find_in_parent_folders("root.hcl")
}

include "state" {
  path = find_in_parent_folders("_managed-state.hcl")
}

include "kubernetes" {
  path = find_in_parent_folders("_kubernetes.hcl")
}

dependencies {
  paths = ["../namespaces"]
}

terraform {
  source = "../_modules/ntfy"
}
