.POSIX:
.PHONY: test lint e2e benchmarks list render-fixture

filter ?= .
override filter := $(value filter)
labels ?=
env ?=
report_dir ?= $(CURDIR)/reports

test:
	go vet ./...
	go test -race -short -count=1 ./...

lint:
	golangci-lint fmt --config ../.golangci.yml --diff ./...
	golangci-lint run --config ../.golangci.yml ./...

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
	E2E=1 BENCHMARKS=1 TEST_ENV='$(env)' go tool ginkgo --race --timeout="$$timeout" \
		--fail-on-empty --focus="$$focus" --label-filter='$(labels)' -v \
		--poll-progress-after=60s --poll-progress-interval=30s \
		--output-dir='$(report_dir)/$@' --json-report=report.json --junit-report=junit.xml ./$@ -- -test.run="^$${entry}\$$"

list:
	E2E=1 BENCHMARKS=1 go tool ginkgo --dry-run -v ./e2e ./benchmarks -- -test.run='^(TestCluster|TestBenchmarks)$$'

fixture_chart ?= app-template
fixture_chart_args ?= --repo https://bjw-s-labs.github.io/helm-charts --version 5.3.0

render-fixture:
	@mkdir -p '$(report_dir)/fixture'
	helm template test $(fixture_chart) $(fixture_chart_args) --namespace test -f ../apps/test.yaml \
		--set global.domain=staging.khuedoan.com --api-versions gateway.networking.k8s.io/v1/HTTPRoute \
		> '$(report_dir)/fixture/manifest.yaml'
