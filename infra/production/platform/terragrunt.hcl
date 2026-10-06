include "root" {
  path = find_in_parent_folders("root.hcl")
}

include "kubernetes" {
  path = find_in_parent_folders("kubernetes.hcl")
}

dependencies {
  paths = ["../argocd"]
}

terraform {
  source = "${find_in_parent_folders("_modules")}//platform"
}

inputs = {
  repository = {
    url      = get_env("GITOPS_REPOSITORY_URL", "https://github.com/khuedoan/homelab")
    revision = get_env("GITOPS_REVISION", "master")
  }
  # OpenBao derives reader roles from the app catalog, so the platform module
  # still receives the list of apps.
  apps = [
    for file in fileset("${get_repo_root()}/apps", "*.yaml") :
    trimsuffix(file, ".yaml")
  ]
}
