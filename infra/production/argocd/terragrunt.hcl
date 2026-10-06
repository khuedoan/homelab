include "root" {
  path = find_in_parent_folders("root.hcl")
}

dependency "cluster" {
  config_path = "../cluster"
}

inputs = {
  kubeconfig = dependency.cluster.outputs.kubeconfig_path
  domain     = "khuedoan.com"
}
