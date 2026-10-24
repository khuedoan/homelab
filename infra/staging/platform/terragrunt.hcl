include "root" {
  path = find_in_parent_folders("root.hcl")
}

dependencies {
  paths = ["../argocd"]
}

dependency "backup" {
  config_path = "../backup"
}

terraform {
  source = "${find_in_parent_folders("_modules")}//platform"
}

inputs = {
  load_balancer_cidr = jsondecode(file("${get_terragrunt_dir()}/../cluster/config.json")).load_balancer_cidr
  backup_location    = dependency.backup.outputs.location
  control_plane_ips  = [for host in jsondecode(file("${get_terragrunt_dir()}/../metal/hosts.json")) : host.ip]
  domain             = "staging.khuedoan.com"
  resource_prefix    = "homelab-staging"
  repository = {
    url      = get_env("GITOPS_REPOSITORY_URL", "http://forgejo-http.forgejo:3000/ops/homelab.git")
    revision = get_env("GITOPS_REVISION", "master")
  }
  ceph_replica_count = 3
}
