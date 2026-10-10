.POSIX:
.PHONY: default infra sync sync-secrets sync-forgejo push-forgejo sso smoke-test test docs fmt
.EXPORT_ALL_VARIABLES:

env ?=
kubeconfig ?= infra/$(env)/kubeconfig.yaml
sso_args = --kubeconfig "$(kubeconfig)" --url "https://auth.$$domain" --bootstrap

default: infra

infra:
	@test -n "$(env)" || { echo 'Usage: make infra env=production (or staging)' >&2; exit 1; }
	cd "infra/$(env)" && terragrunt run --all apply
	$(MAKE) sync

sync:
	@test -n "$(env)" || { echo 'Usage: make sync env=production (or staging)' >&2; exit 1; }
	$(MAKE) sync-secrets
	$(MAKE) sync-forgejo
	$(MAKE) push-forgejo
	$(MAKE) sso

sync-secrets:
	@test -n "$(env)" || { echo 'Usage: make sync-secrets env=production (or staging)' >&2; exit 1; }
	toolbox infra secrets sync --environment "$(env)" --kubeconfig "$(kubeconfig)"

sync-forgejo:
	@test -n "$(env)" || { echo 'Usage: make sync-forgejo env=production (or staging)' >&2; exit 1; }
	toolbox infra forgejo sync --kubeconfig "$(kubeconfig)" --domain "$$(cd infra/$(env)/platform && terragrunt render --json | jq -r '.inputs.domain')"

push-forgejo:
	@test -n "$(env)" || { echo 'Usage: make push-forgejo env=production (or staging)' >&2; exit 1; }
	toolbox infra forgejo push --kubeconfig "$(kubeconfig)" --domain "$$(cd infra/$(env)/platform && terragrunt render --json | jq -r '.inputs.domain')"

sso:
	@test -n "$(env)" || { echo 'Usage: make sso env=production (or staging)' >&2; exit 1; }
	@set -eu; \
	domain=$$(cd infra/$(env)/platform && terragrunt render --json | jq -er '.inputs.domain'); \
	toolbox sso $(sso_args) client ensure grafana --display-name Grafana --redirect-uri "https://grafana.$$domain/login/generic_oauth"; \
	toolbox sso $(sso_args) secrets sync

smoke-test:
	make -C tests e2e filter='^Apps$$' env='$(env)'

test:
	make -C toolbox test
	make -C tests test

docs:
	mkdocs serve

fmt:
	treefmt
