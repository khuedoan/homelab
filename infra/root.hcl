terraform_binary = "tofu"

terraform {
  before_hook "unsupported_destroy" {
    commands = ["destroy"]
    execute  = ["sh", "${get_repo_root()}/infra/pending", "Infrastructure teardown"]
  }
}
