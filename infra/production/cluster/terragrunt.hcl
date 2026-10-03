include "root" {
  path = find_in_parent_folders("root.hcl")
}

dependencies {
  paths = ["../metal"]
}

terraform {
  source = "${find_in_parent_folders("_modules")}//cluster"

  after_hook "kubeconfig" {
    commands = ["apply"]
    execute  = ["sh", "-c", "cd \"$1\" && exec toolbox cluster kubeconfig --environment production --output infra/kubeconfig.yaml", "kubeconfig", get_repo_root()]
  }
}

inputs = {
  kubeconfig = "${get_repo_root()}/infra/kubeconfig.yaml"
}
