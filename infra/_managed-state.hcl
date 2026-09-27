generate "backend" {
  path      = "backend.tf"
  if_exists = "overwrite_terragrunt"
  contents  = <<-EOF
    terraform {
      backend "remote" {
        hostname     = "app.terraform.io"
        organization = "khuedoan"
        workspaces {
          name = "homelab-infra-${path_relative_to_include("state")}"
        }
      }
    }
  EOF
}

terraform {
  before_hook "pending_state_migration" {
    commands = ["apply"]
    execute  = ["sh", "${get_repo_root()}/infra/pending", "Managed-service state migration and remote workspace setup for local execution"]
  }
}
