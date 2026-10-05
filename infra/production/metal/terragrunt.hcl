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
    execute  = ["./nixie", get_working_dir(), "${get_terragrunt_dir()}/hosts.json"]
  }

  after_hook "enroll" {
    commands = ["apply"]
    execute  = ["sh", "-c", "cd \"$1\" && exec toolbox cluster enroll --environment production", "enroll", get_repo_root()]
  }
}

inputs = {
  hosts   = local.hosts
  cluster = jsondecode(file("${get_terragrunt_dir()}/../cluster/config.json"))
}
