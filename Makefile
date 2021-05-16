.POSIX:

default: run

.PHONY: run
run:
	make -C metal
	make -C infra
	make -C apps

.PHONY: tools
tools:
	make -C tools

.PHONY: docs
docs:
	make -C docs
