default:
    @just --list

mod test "tests/justfile"
mod infra "infra/justfile"

deploy target: (infra::deploy target)

[positional-arguments]
@user action target username *arguments:
    cd infra && python3 src/scripts/user.py "$@"

docs:
    mdbook-mermaid install docs
    mdbook serve docs --port 8000

publish:
    #!/usr/bin/env sh
    set -eu
    mdbook-mermaid install docs
    mdbook build docs
    test -e .worktrees/pages/.git || git worktree add .worktrees/pages pages
    cd .worktrees/pages
    test -z "$(git status --porcelain)"
    git pull --ff-only origin pages
    git rm -r --quiet . ':!CNAME'
    cp -a ../../docs/book/. .
    git add --all --force
    git diff --cached --quiet || git commit -m "docs: publish documentation"
    git push origin HEAD:pages

lint: test::lint infra::lint

fmt: test::fmt infra::fmt
    nix fmt
    ruff format .
    just --fmt
