.POSIX:
.PHONY: *
.EXPORT_ALL_VARIABLES:

KUBECONFIG = $(shell pwd)/metal/kubeconfig.yaml
KUBE_CONFIG_PATH = $(KUBECONFIG)
PXE_INTERFACE ?= $(shell ip -4 route show default | awk '/default/ { print $$5; exit }')
PXE_ADDRESS := $(shell ip -4 -o address show dev $(PXE_INTERFACE) scope global | awk '{ sub(/\/.*/, "", $$4); print $$4 }')
SSH_KEY := ${HOME}/.ssh/id_ed25519
INSTALL_KEY := $(shell grep -o 'ssh-ed25519 AAAA[A-Za-z0-9+/=]*' metal/installer.nix | head -1)

default: metal system external smoke-test post-install fmt

configure:
	./scripts/configure
	git status

metal:
	@test -n "${PXE_INTERFACE}" || { \
		echo "no default-route interface for PXE, set PXE_INTERFACE" >&2; \
		exit 1; \
	}
	@test -n "${PXE_ADDRESS}" || { \
		echo "${PXE_INTERFACE} has no IPv4 address" >&2; \
		exit 1; \
	}
	@test "$$(ssh-keygen -y -f '${SSH_KEY}')" = "${INSTALL_KEY}" || { \
		echo "${SSH_KEY} does not match the authorized key in metal/installer.nix" >&2; \
		exit 1; \
	}
	sudo env "PATH=$$PATH" nixie \
		--address "${PXE_ADDRESS}" \
		--installer .#nixosConfigurations.installer \
		--flake . \
		--hosts metal/hosts.json \
		--install-ssh-key "${SSH_KEY}" \
		--deployment-ssh-key "${SSH_KEY}"

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
