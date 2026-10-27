include "root" {
  path = find_in_parent_folders("root.hcl")
}

terraform {
  source = "${find_in_parent_folders("_modules")}/nixos//"

  after_hook "nixie" {
    commands    = ["apply"]
    working_dir = "${get_working_dir()}/flake"
    execute = [
      "sudo", "env", "PATH=${get_env("PATH")}", "nixie",
      "--installer", "path:.#nixosConfigurations.installer",
      "--flake", "path:.",
      "--hosts", "${get_terragrunt_dir()}/hosts.json",
      "--ssh-agent-socket", get_env("SSH_AUTH_SOCK", ""),
    ]
  }

  after_hook "enroll" {
    commands    = ["apply"]
    working_dir = get_repo_root()
    execute     = ["lab", "cluster", "enroll", "--environment", "production"]
  }

  after_hook "kubeconfig" {
    commands    = ["apply"]
    working_dir = get_repo_root()
    execute     = ["lab", "cluster", "kubeconfig", "--environment", "production", "--output", "infra/production/kubeconfig.yaml"]
  }
}

inputs = {
  hosts   = jsondecode(file("${get_terragrunt_dir()}/hosts.json"))
  cluster = jsondecode(file("${get_terragrunt_dir()}/../cluster/config.json"))
}
