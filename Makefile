.POSIX:
.PHONY: *
.EXPORT_ALL_VARIABLES:

KUBECONFIG = $(shell pwd)/infra/kubeconfig.yaml
KUBE_CONFIG_PATH = $(KUBECONFIG)
ENV ?= production
HOSTS_FILE := infra/$(ENV)/metal/hosts.json
NIXOS_HOSTS := infra/nixos/hosts.json

default: metal fmt

$(NIXOS_HOSTS): $(HOSTS_FILE)
	cp $(HOSTS_FILE) $(NIXOS_HOSTS)

metal: $(NIXOS_HOSTS)
	./infra/_modules/nixos/nixie "$(CURDIR)" "$(HOSTS_FILE)"

external:
	make -C external

smoke-test:
	make -C tests e2e filter='^Apps$$' config='$(TEST_CONFIG)'

post-install:
	toolbox integrations setup

# TODO maybe there's a better way to manage backup with GitOps?
backup:
	toolbox backup setup --namespace=actualbudget --pvc=actualbudget-data
	toolbox backup setup --namespace=jellyfin --pvc=jellyfin-data

restore:
	toolbox backup restore --namespace=actualbudget --pvc=actualbudget-data
	toolbox backup restore --namespace=jellyfin --pvc=jellyfin-data

test:
	make -C toolbox test
	make -C tests test

docs:
	mkdocs serve

fmt:
	treefmt
