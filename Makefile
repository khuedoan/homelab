.POSIX:
.PHONY: default infra sync sync-secrets sync-forgejo push-forgejo sso smoke-test test lint check graph e2e benchmarks list render-fixture docs fmt
.EXPORT_ALL_VARIABLES:

env ?=
filter ?= .
override filter := $(value filter)
labels ?=
report_dir ?= $(CURDIR)/tests/reports
fixture_chart ?= app-template
fixture_chart_args ?= --repo https://bjw-s-labs.github.io/helm-charts --version 5.3.0
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
	lab infra secrets sync --environment "$(env)" --kubeconfig "$(kubeconfig)"

sync-forgejo:
	@test -n "$(env)" || { echo 'Usage: make sync-forgejo env=production (or staging)' >&2; exit 1; }
	lab infra forgejo sync --kubeconfig "$(kubeconfig)" --domain "$$(cd infra/$(env)/platform && terragrunt render --json | jq -r '.inputs.domain')"

push-forgejo:
	@test -n "$(env)" || { echo 'Usage: make push-forgejo env=production (or staging)' >&2; exit 1; }
	lab infra forgejo push --kubeconfig "$(kubeconfig)" --domain "$$(cd infra/$(env)/platform && terragrunt render --json | jq -r '.inputs.domain')"

# TODO remove --rs256 once Actual Budget supports ES256 https://github.com/actualbudget/actual/issues/6524
sso:
	@test -n "$(env)" || { echo 'Usage: make sso env=production (or staging)' >&2; exit 1; }
	@set -eu; \
	domain=$$(cd infra/$(env)/platform && terragrunt render --json | jq -er '.inputs.domain'); \
	lab sso $(sso_args) group ensure users; \
	lab sso $(sso_args) group ensure developers; \
	lab sso $(sso_args) group ensure admins; \
	lab sso $(sso_args) client ensure grafana --display-name Grafana --redirect-uri "https://grafana.$$domain/login/generic_oauth"; \
	lab sso $(sso_args) client ensure forgejo --display-name Forgejo --redirect-uri "https://git.$$domain/user/oauth2/SSO/callback"; \
	lab sso $(sso_args) client ensure argocd --display-name 'Argo CD' --redirect-uri "https://argocd.$$domain/auth/callback"; \
	lab sso $(sso_args) client ensure actualbudget --rs256 --display-name 'Actual Budget' --redirect-uri "https://budget.$$domain/openid/callback"; \
	lab sso $(sso_args) client ensure matrix --display-name Matrix --redirect-uri "https://matrix.$$domain/_matrix/client/unstable/login/sso/callback/matrix"; \
	lab sso $(sso_args) client ensure immich --display-name Immich --redirect-uri "https://photos.$$domain/auth/login" --redirect-uri "https://photos.$$domain/user-settings" --redirect-uri "https://photos.$$domain/api/oauth/mobile-redirect" --claim immich_role:users=user --claim immich_role:admins=admin; \
	lab sso $(sso_args) client ensure kiali --group admins --rs256 --display-name Kiali --redirect-uri "https://kiali.$$domain/"; \
	lab sso $(sso_args) client ensure wireguard --display-name 'WireGuard Portal' --redirect-uri "https://vpn.$$domain/api/v0/auth/login/sso/callback"; \
	lab sso $(sso_args) client ensure opencloud --display-name OpenCloud --public --allow-localhost-redirects --redirect-uri "https://drive.$$domain/" --redirect-uri "https://drive.$$domain/oidc-callback.html" --redirect-uri "https://drive.$$domain/oidc-silent-redirect.html" --redirect-uri "oc://android.opencloud.eu" --redirect-uri "oc://ios.opencloud.eu"; \
	lab sso $(sso_args) client ensure zot --display-name Zot --public --redirect-uri "https://registry.$$domain/zot/auth/callback/oidc"; \
	lab sso $(sso_args) client ensure openbao --group admins --display-name OpenBao --redirect-uri "https://openbao.$$domain/ui/vault/auth/oidc/oidc/callback"; \
	lab sso $(sso_args) secrets sync; \
	lab sso $(sso_args) openbao configure --redirect-uri "https://openbao.$$domain/ui/vault/auth/oidc/oidc/callback"

smoke-test:
	$(MAKE) e2e filter='^Apps$$' env='$(env)'

test:
	cd cli && go vet ./...
	cd cli && go test -race ./...
	cd tests && go vet ./...
	cd tests && go test -race -short -count=1 ./...

lint:
	cd cli && golangci-lint fmt --config ../.golangci.yml --diff ./...
	cd cli && golangci-lint run --config ../.golangci.yml ./...
	cd tests && golangci-lint fmt --config ../.golangci.yml --diff ./...
	cd tests && golangci-lint run --config ../.golangci.yml ./...

check:
	cd infra && terragrunt hcl validate
	tofu fmt -check -recursive infra

graph:
	@test -n "$(env)" || { echo 'Usage: make graph env=production (or staging)' >&2; exit 1; }
	cd "infra/$(env)" && terragrunt dag graph

e2e benchmarks:
	@test -n "$(env)" || { echo 'Usage: make $@ env=staging' >&2; exit 1; }
	@focus='$(filter)'; \
	case "$$focus" in \
	  '^Networking$$'|'^NetworkPolicy$$'|'^LoadBalancer$$'|'^Storage$$'|'^Snapshots$$'|'^Backups$$'|'^Registry$$'|'^GitOps$$'|'^Secrets$$'|'^Apps$$'|'^Security$$') focus="$${focus%?} " ;; \
	esac; \
	case '$@' in \
	  e2e) timeout=30m; entry=TestCluster ;; \
	  benchmarks) timeout=60m; entry=TestBenchmarks ;; \
	esac; \
	cd tests && E2E=1 BENCHMARKS=1 TEST_ENV='$(env)' go tool ginkgo --race --timeout="$$timeout" \
		--fail-on-empty --focus="$$focus" --label-filter='$(labels)' -v \
		--poll-progress-after=60s --poll-progress-interval=30s \
		--output-dir='$(report_dir)/$@' --json-report=report.json --junit-report=junit.xml ./$@ -- -test.run="^$${entry}\$$"

list:
	cd tests && E2E=1 BENCHMARKS=1 go tool ginkgo --dry-run -v ./e2e ./benchmarks -- -test.run='^(TestCluster|TestBenchmarks)$$'

render-fixture:
	@mkdir -p '$(report_dir)/fixture'
	helm template test $(fixture_chart) $(fixture_chart_args) --namespace test -f apps/test.yaml \
		--set global.domain=staging.khuedoan.com --api-versions gateway.networking.k8s.io/v1/HTTPRoute \
		> '$(report_dir)/fixture/manifest.yaml'

docs:
	mkdocs serve

fmt:
	treefmt
