include "root" {
  path = find_in_parent_folders("root.hcl")
}

locals {
  hosts = jsondecode(file("${get_terragrunt_dir()}/hosts.json"))
}

terraform {
  source = "${find_in_parent_folders("_modules")}//nixos"

  after_hook "nixie" {
    commands = ["apply"]
    execute  = ["${get_repo_root()}/infra/_modules/nixos/nixie", "${get_repo_root()}", "${get_terragrunt_dir()}/hosts.json"]
  }
}

inputs = {
  flake = "${get_repo_root()}/infra/nixos"
  hosts = local.hosts
}
