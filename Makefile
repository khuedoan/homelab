.POSIX:
.PHONY: *
.EXPORT_ALL_VARIABLES:

env ?= dev
KUBECONFIG = $(shell pwd)/metal/kubeconfig-${env}.yaml
KUBE_CONFIG_PATH = $(KUBECONFIG)

default: metal system external smoke-test post-install

configure:
	./scripts/configure
	git status

metal:
	[ "$(env)" = "dev" ] \
		&& make k3d \
		|| make -C metal

system:
	make -C system

external:
	[ "$(env)" != "dev" ] && make -C external

smoke-test:
	make -C test filter=Smoke

post-install:
	@./scripts/hacks

# TODO maybe there's a better way to manage backup with GitOps?
backup:
	./scripts/backup --action setup --namespace=actualbudget --pvc=actualbudget-data
	./scripts/backup --action setup --namespace=jellyfin --pvc=jellyfin-data

restore:
	./scripts/backup --action restore --namespace=actualbudget --pvc=actualbudget-data
	./scripts/backup --action restore --namespace=jellyfin --pvc=jellyfin-data

test:
	make -C test

clean:
	docker compose --project-directory ./metal/roles/pxe_server/files down
	k3d cluster delete homelab-dev

docs:
	mkdocs serve

git-hooks:
	pre-commit install

info:
	kubectl cluster-info

k3d:
	k3d cluster start homelab-dev || k3d cluster create --config metal/k3d-${env}.yaml
	k3d kubeconfig get homelab-dev > metal/kubeconfig-${env}.yaml
