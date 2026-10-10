terraform {
  before_hook "verify_cluster" {
    commands = ["plan", "apply", "refresh", "import"]
    execute = [
      "sh", "-c",
      "cd \"$1\" && exec toolbox cluster kubeconfig --environment production --output infra/production/kubeconfig.yaml",
      "verify-cluster", get_repo_root(),
    ]
  }
}

inputs = {
  kubeconfig = "${get_repo_root()}/infra/production/kubeconfig.yaml"
}
