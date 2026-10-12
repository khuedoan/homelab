include "root" {
  path = find_in_parent_folders("root.hcl")
}

dependencies {
  paths = ["../argocd"]
}

terraform {
  source = "${find_in_parent_folders("_modules")}//platform"
}

inputs = {
  domain          = "staging.khuedoan.com"
  resource_prefix = "homelab-staging"
  repository = {
    url      = get_env("GITOPS_REPOSITORY_URL", "http://forgejo-http.forgejo:3000/ops/homelab.git")
    revision = get_env("GITOPS_REVISION", "master")
  }
  ceph_replica_count = 3

  # TODO remove these workload overrides when staging has enough resources
  apps                   = []
  log_collection_enabled = false
}
