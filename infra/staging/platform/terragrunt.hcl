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
    url      = get_env("GITOPS_REPOSITORY_URL", "http://forgejo-http.forgejo:3000/forgejo_admin/staging-apps.git")
    revision = get_env("GITOPS_REVISION", "main")
  }
  apps = ["pairdrop"]
}
