include "root" {
  path = find_in_parent_folders("root.hcl")
}

dependencies {
  paths = ["../metal"]
}

locals {
  cluster = jsondecode(file("${get_terragrunt_dir()}/config.json"))
}

inputs = {
  hosts       = jsondecode(file("${get_terragrunt_dir()}/../metal/hosts.json"))
  initializer = local.cluster.init_host
  vip         = local.cluster.vip
  kubeconfig  = "${get_terragrunt_dir()}/../kubeconfig.yaml"
}
