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
  apps               = ["pairdrop"]
  ceph_replica_count = 1
}
