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
	# sudo nix run .#nixosPxeServer
	# echo 'waiting until installer is booted' && sleep 30
	# ssh root@192.168.1.2 cat /etc/os-release | grep VARIANT_ID=installer && nixos-anywhere --flake .#metal1 --target-host root@192.168.1.2
	# nixos-rebuild --flake . --target-host root@metal1.local switch

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

docs:
	mkdocs serve

git-hooks:
	pre-commit install

fmt:
	treefmt
	cd tools && go fmt ./...
