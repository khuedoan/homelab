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
  domain          = "khuedoan.com"
  resource_prefix = "homelab-production"
  repository = {
    url      = get_env("GITOPS_REPOSITORY_URL", "http://forgejo-http.forgejo:3000/ops/homelab.git")
    revision = get_env("GITOPS_REVISION", "master")
  }
  # OpenBao derives reader roles from the app catalog, so the platform module
  # still receives the list of apps.
  apps = [
    for file in fileset("${get_repo_root()}/apps", "*.yaml") :
    trimsuffix(file, ".yaml")
  ]
}
