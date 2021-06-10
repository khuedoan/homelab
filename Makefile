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
	make -C metal lint
	make -C infra lint
