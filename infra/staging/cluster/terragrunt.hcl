include "root" {
  path = find_in_parent_folders("root.hcl")
}

dependencies {
  paths = ["../metal", "../hetzner-metal"]
}

terraform {
  source = "../_modules/command"

  after_hook "ready" {
    commands = ["apply"]
    # TODO
    # execute  = ["sh", "${get_repo_root()}/infra/getkubectlsomething", "Cluster readiness and kubeconfig export"]
  }
}
