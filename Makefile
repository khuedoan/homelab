.POSIX:
.PHONY: *
.EXPORT_ALL_VARIABLES:

KUBECONFIG = $(shell pwd)/metal/kubeconfig.yaml
KUBE_CONFIG_PATH = $(KUBECONFIG)

default: metal bootstrap external wait

configure:
	./scripts/configure
	git status

metal:
	make -C metal

bootstrap:
	make -C bootstrap

external:
	make -C external

wait:
	./scripts/wait-main-apps

tools:
	make -C tools

dev:
	make -C metal cluster env=dev
	make -C bootstrap

docs:
	docker run \
		--rm \
		--interactive \
		--tty \
		--publish 8000:8000 \
		--volume $(shell pwd):/docs \
		squidfunk/mkdocs-material

git-hooks:
	pre-commit install
