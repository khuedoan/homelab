.POSIX:
.PHONY: *
.EXPORT_ALL_VARIABLES:

KUBECONFIG = $(shell pwd)/metal/kubeconfig.yaml
KUBE_CONFIG_PATH = $(KUBECONFIG)

default: metal bootstrap wait

all: metal bootstrap external wait

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

docs:
	make -C docs

dev:
	make -C metal cluster env=dev
	make -C bootstrap
