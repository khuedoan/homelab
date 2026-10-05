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
  apps = { for chart in fileset("${get_repo_root()}/apps", "*/Chart.yaml") : dirname(chart) => {
    path      = "apps/${dirname(chart)}"
    namespace = dirname(chart)
    values    = ""
  } }
}
