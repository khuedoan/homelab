include "root" {
  path = find_in_parent_folders("root.hcl")
}

dependencies {
  paths = ["../argocd"]
}

dependency "cluster" {
  config_path = "../cluster"
}

dependency "backup" {
  config_path = "../backup"
}

inputs = {
  kubeconfig         = dependency.cluster.outputs.kubeconfig_path
  load_balancer_cidr = jsondecode(file("${get_terragrunt_dir()}/../cluster/config.json")).load_balancer_cidr
  backup_location    = dependency.backup.outputs.location
  control_plane_ips  = [for host in jsondecode(file("${get_terragrunt_dir()}/../metal/hosts.json")) : host.ip]
  domain             = "staging.khuedoan.com"
  ceph_replica_count = 3
  apps = [for file in fileset("${get_repo_root()}/apps", "*.yaml") : trimsuffix(file, ".yaml")]
}
