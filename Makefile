.POSIX:
.PHONY: *
.EXPORT_ALL_VARIABLES:

KUBECONFIG = $(shell pwd)/metal/kubeconfig.yaml
KUBE_CONFIG_PATH = $(KUBECONFIG)

default: metal bootstrap external smoke-test post-install clean

configure:
	./scripts/configure
	git status

metal:
	make -C metal

bootstrap:
	make -C bootstrap

external:
	make -C external

smoke-test:
	make -C test filter=Smoke

post-install:
	@./scripts/hacks

tools:
	@docker run \
		--rm \
		--interactive \
		--tty \
		--network host \
		--env "KUBECONFIG=${KUBECONFIG}" \
		--volume "/var/run/docker.sock:/var/run/docker.sock" \
		--volume $(shell pwd):$(shell pwd) \
		--volume ${HOME}/.ssh:/root/.ssh \
		--volume ${HOME}/.terraform.d:/root/.terraform.d \
		--volume homelab-tools-cache:/root/.cache \
		--volume homelab-tools-nix:/nix \
		--workdir $(shell pwd) \
		--entrypoint /bin/sh \
		docker.io/nixos/nix -c "\
		git config --global --add safe.directory $(shell pwd) && \
		nix --experimental-features 'nix-command flakes' develop"

test:
	make -C test

clean:
	docker compose --project-directory ./metal/roles/pxe_server/files down

dev:
	make -C metal cluster env=dev
	make -C bootstrap

docs:
	mkdocs serve

git-hooks:
	pre-commit install
