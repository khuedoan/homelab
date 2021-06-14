.POSIX:

default: apply

.PHONY: metal
metal:
	make -C metal

.PHONY: infra
infra:
	make -C infra

.PHONY: apps
apps:
	make -C apps

.PHONY: tools
tools:
	make -C tools

.PHONY: docs
docs:
	make -C docs

.PHONY: apply
apply: metal infra apps

lint:
	# TODO (feature) Add lint checks for everything
	make -C metal lint
	make -C infra lint

hooks:
	cp ./scripts/hooks/* .git/hooks/
