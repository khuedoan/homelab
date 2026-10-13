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
	$(MAKE) sso
	$(MAKE) sync-forgejo
	$(MAKE) push-forgejo

sync-secrets:
	@test -n "$(env)" || { echo 'Usage: make sync-secrets env=production (or staging)' >&2; exit 1; }
	toolbox infra secrets sync --environment "$(env)" --kubeconfig "$(kubeconfig)"

sync-forgejo:
	@test -n "$(env)" || { echo 'Usage: make sync-forgejo env=production (or staging)' >&2; exit 1; }
	toolbox infra forgejo sync --kubeconfig "$(kubeconfig)" --domain "$$(cd infra/$(env)/platform && terragrunt render --json | jq -r '.inputs.domain')"

push-forgejo:
	@test -n "$(env)" || { echo 'Usage: make push-forgejo env=production (or staging)' >&2; exit 1; }
	toolbox infra forgejo push --kubeconfig "$(kubeconfig)" --domain "$$(cd infra/$(env)/platform && terragrunt render --json | jq -r '.inputs.domain')"

# TODO remove --rs256 once Actual Budget supports ES256 https://github.com/actualbudget/actual/issues/6524
sso:
	@test -n "$(env)" || { echo 'Usage: make sso env=production (or staging)' >&2; exit 1; }
	@set -eu; \
	domain=$$(cd infra/$(env)/platform && terragrunt render --json | jq -er '.inputs.domain'); \
	toolbox sso $(sso_args) client ensure grafana --display-name Grafana --redirect-uri "https://grafana.$$domain/login/generic_oauth"; \
	toolbox sso $(sso_args) client ensure forgejo --display-name Forgejo --redirect-uri "https://git.$$domain/user/oauth2/SSO/callback"; \
	toolbox sso $(sso_args) client ensure argocd --display-name 'Argo CD' --redirect-uri "https://argocd.$$domain/auth/callback"; \
	toolbox sso $(sso_args) client ensure actualbudget --rs256 --display-name 'Actual Budget' --redirect-uri "https://budget.$$domain/openid/callback"; \
	toolbox sso $(sso_args) client ensure paperless --display-name Paperless --redirect-uri "https://paperless.$$domain/accounts/oidc/sso/login/callback/"; \
	toolbox sso $(sso_args) client ensure opencloud --display-name OpenCloud --public --redirect-uri "https://drive.$$domain/" --redirect-uri "https://drive.$$domain/oidc-callback.html" --redirect-uri "https://drive.$$domain/oidc-silent-redirect.html"; \
	toolbox sso $(sso_args) client ensure zot --display-name Zot --public --redirect-uri "https://registry.$$domain/zot/auth/callback/oidc"; \
	toolbox sso $(sso_args) client ensure openbao --display-name OpenBao --redirect-uri "https://openbao.$$domain/ui/vault/auth/oidc/oidc/callback"; \
	toolbox sso $(sso_args) secrets sync; \
	toolbox sso $(sso_args) openbao configure --redirect-uri "https://openbao.$$domain/ui/vault/auth/oidc/oidc/callback"

smoke-test:
	make -C tests e2e filter='^Apps$$' env='$(env)'

test:
	make -C toolbox test
	make -C tests test

docs:
	mkdocs serve

fmt:
	treefmt
