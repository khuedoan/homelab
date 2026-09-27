include "root" {
  path = find_in_parent_folders("root.hcl")
}

include "state" {
  path = find_in_parent_folders("_local-state.hcl")
}

terraform {
  source = "../_modules/command"

  after_hook "nixie" {
    commands = ["apply"]
    execute  = ["sh", "${get_repo_root()}/infra/pending", "Nixie install-or-update for metal nodes"]
  }
}
