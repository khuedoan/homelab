.POSIX:
.PHONY: *
.EXPORT_ALL_VARIABLES:

KUBECONFIG = $(shell pwd)/metal/kubeconfig.yaml
KUBE_CONFIG_PATH = $(KUBECONFIG)

default: metal system external smoke-test post-install clean fmt

configure:
	./scripts/configure
	git status

metal:
	make -C metal
	# TODO new install flow
	# sudo -v
	# sudo nix run .#nixos-pxe
	# echo 'waiting until installer is booted' && sleep 30
	# nixos-anywhere --flake .#metal1 --target-host root@192.168.1.6
	# nixos-rebuild --flake .#metal1 --target-host root@192.168.1.6 switch

system:
	make -C system

external:
	make -C external

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

docs:
	mkdocs serve

git-hooks:
	pre-commit install

fmt:
	treefmt
