generate "kubernetes_provider" {
  path      = "provider.tf"
  if_exists = "overwrite_terragrunt"
  contents  = <<-EOF
    provider "kubernetes" {
      config_path = "${get_repo_root()}/infra/kubeconfig.yaml"
    }
  EOF
}
