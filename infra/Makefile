.POSIX:
.PHONY: default check graph apply

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
