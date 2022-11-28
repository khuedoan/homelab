.POSIX:
.PHONY: *
.EXPORT_ALL_VARIABLES:

KUBECONFIG = $(shell pwd)/metal/kubeconfig.yaml
KUBE_CONFIG_PATH = $(KUBECONFIG)

.DEFAULT: help
help:	## show this help menu.
	@echo "Usage: make [TARGET ...]"
	@echo ""
	@@grep -hE "#[#]" $(MAKEFILE_LIST) | sed -e 's/\\$$//' | awk 'BEGIN {FS = "[:=].*?#[#] "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

metal:	## run ansible to configure k3s
metal:
	make -C metal

bootstrap:	## deploy ArgoCD configured to sync with this repo
bootstrap:
	make -C bootstrap

dev:	## deploy k3s
dev:
	test/step_by_step_tests.sh

docs:	## run doc server in local at port 8000
docs:
	docker run \
		--rm \
		--interactive \
		--tty \
		--publish 8000:8000 \
		--volume $(shell pwd):/docs \
		squidfunk/mkdocs-material

git-hooks:	## pre-commit install
git-hooks:
	pre-commit install
