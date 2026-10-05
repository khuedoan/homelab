.POSIX:
.PHONY: default check graph apply repos charts

KUBECONFIG ?= $(CURDIR)/kubeconfig.yaml
KUBE_CONFIG_PATH ?= $(KUBECONFIG)

.EXPORT_ALL_VARIABLES:

default: check

check:
	terragrunt hcl validate
	tofu fmt -check -recursive .

graph:
	terragrunt dag graph

apply:
	terragrunt run --all apply

repos:
	@set -eu; for url in $$(awk '$$1 == "repository:" && $$2 !~ /^oci:/ {print $$2}' _modules/platform/*/Chart.yaml | sort -u); do \
		name=$$(printf '%s' "$$url" | cksum | cut -d ' ' -f 1); \
		helm repo add "infra-$$name" "$$url"; \
	done

charts: repos
	@set -eu; for chart in _modules/platform/*/Chart.yaml; do \
		helm dependency build --skip-refresh "$${chart%/Chart.yaml}"; \
	done
