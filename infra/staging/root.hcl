locals {
  # Metal and cluster use the default local backend in the Terragrunt cache.
  stateless = contains(["metal", "cluster"], path_relative_to_include())
}

generate "backend" {
  path              = "backend.tf.json"
  if_exists         = "overwrite"
  disable           = local.stateless
  if_disabled       = "remove_terragrunt"
  disable_signature = true
  contents = jsonencode({
    terraform = {
      backend = {
        local = {
          path = "${get_parent_terragrunt_dir()}/.state/${path_relative_to_include()}.tfstate"
        }
      }
    }
  })
}

terraform {
  before_hook "unsupported_destroy" {
    commands = ["destroy"]
    execute  = ["sh", "-c", "echo 'Infrastructure teardown is unsupported' >&2; exit 1"]
  }
}
