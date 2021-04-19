.POSIX:

default: tools

.PHONY: tools
tools:
	make -C tools

.PHONY: metal
metal:
	make -C metal

.PHONY: infra
infra:
	make -C infra

.PHONY: apps
apps:
	make -C apps

.PHONY: docs
docs:
	make -C docs
