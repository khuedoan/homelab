default:
    @just --list

mod test "tests/justfile"
mod infra "infra/justfile"

deploy target: (infra::deploy target)

user target username display_name email groups="": (infra::user target username display_name email groups)

docs:
    mdbook-mermaid install docs
    mdbook serve docs --port 8000

lint: test::lint infra::lint

fmt: test::fmt infra::fmt
    nix fmt
    ruff format .
    just --fmt
